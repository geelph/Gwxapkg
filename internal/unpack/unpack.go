package unpack

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/25smoking/Gwxapkg/internal/key"
	"github.com/25smoking/Gwxapkg/internal/scanner"

	"github.com/25smoking/Gwxapkg/internal/config"

	formatter2 "github.com/25smoking/Gwxapkg/internal/formatter"
)

const (
	maxFileCount      = 102400
	maxFileNameLength = 1024
	maxSingleFileSize = 128 * 1024 * 1024
)

const (
	stageHeaderValidation = "头部校验"
	stageIndexAnalysis    = "索引分析"
	stagePathPlanning     = "路径规划"
	stageFileRead         = "文件读取"
	stageFileFormat       = "文件格式化"
	stageFileWrite        = "文件写入"
	stageSensitiveScan    = "敏感扫描"
)

var errDuplicatePlannedPath = errors.New("duplicate planned path")

type WxapkgFile struct {
	NameLen uint32
	Name    string
	Offset  uint32
	Size    uint32
}

type plannedFile struct {
	Index        int
	EntryName    string
	RelativePath string
	FullPath     string
	Offset       uint32
	Size         uint32
}

type packagePlan struct {
	SourcePath string
	OutputDir  string
	FileNames  []string
	Files      []plannedFile
}

type packageStageError struct {
	SourcePath string
	Stage      string
	File       string
	Err        error
}

func (e *packageStageError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("wxapkg=%s 阶段=%s 文件=%s: %v", e.SourcePath, e.Stage, e.File, e.Err)
	}
	return fmt.Sprintf("wxapkg=%s 阶段=%s: %v", e.SourcePath, e.Stage, e.Err)
}

func (e *packageStageError) Unwrap() error {
	return e.Err
}

// UnpackWxapkg 解包 wxapkg 文件并将内容保存到指定目录。
func UnpackWxapkg(data []byte, sourcePath string, outputDir string) ([]string, error) {
	plan, err := analyzePackage(data, sourcePath, outputDir)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(data)
	if err := writePlannedFiles(plan, reader); err != nil {
		return nil, err
	}

	return plan.FileNames, nil
}

func analyzePackage(data []byte, sourcePath string, outputDir string) (*packagePlan, error) {
	reader := bytes.NewReader(data)

	outputAbs, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, wrapStageError(sourcePath, stagePathPlanning, "", fmt.Errorf("解析输出目录失败: %w", err))
	}

	var firstMark byte
	if err := binary.Read(reader, binary.BigEndian, &firstMark); err != nil {
		return nil, wrapStageError(sourcePath, stageHeaderValidation, "", fmt.Errorf("读取首标记失败: %w", err))
	}
	if firstMark != 0xBE {
		return nil, wrapStageError(sourcePath, stageHeaderValidation, "", fmt.Errorf("无效的 wxapkg 文件: 首标记不正确"))
	}

	var info1, indexInfoLength, bodyInfoLength uint32
	if err := binary.Read(reader, binary.BigEndian, &info1); err != nil {
		return nil, wrapStageError(sourcePath, stageHeaderValidation, "", fmt.Errorf("读取 info1 失败: %w", err))
	}
	if err := binary.Read(reader, binary.BigEndian, &indexInfoLength); err != nil {
		return nil, wrapStageError(sourcePath, stageHeaderValidation, "", fmt.Errorf("读取索引段长度失败: %w", err))
	}
	if err := binary.Read(reader, binary.BigEndian, &bodyInfoLength); err != nil {
		return nil, wrapStageError(sourcePath, stageHeaderValidation, "", fmt.Errorf("读取数据段长度失败: %w", err))
	}

	if uint64(indexInfoLength)+uint64(bodyInfoLength) > uint64(len(data)) {
		return nil, wrapStageError(sourcePath, stageHeaderValidation, "", fmt.Errorf(
			"文件长度不足: 索引段(%d) + 数据段(%d) > 文件总长度(%d)",
			indexInfoLength, bodyInfoLength, len(data),
		))
	}

	var lastMark byte
	if err := binary.Read(reader, binary.BigEndian, &lastMark); err != nil {
		return nil, wrapStageError(sourcePath, stageHeaderValidation, "", fmt.Errorf("读取尾标记失败: %w", err))
	}
	if lastMark != 0xED {
		return nil, wrapStageError(sourcePath, stageHeaderValidation, "", fmt.Errorf("无效的 wxapkg 文件: 尾标记不正确"))
	}

	var fileCount uint32
	if err := binary.Read(reader, binary.BigEndian, &fileCount); err != nil {
		return nil, wrapStageError(sourcePath, stageIndexAnalysis, "", fmt.Errorf("读取文件数量失败: %w", err))
	}
	if fileCount > maxFileCount {
		return nil, wrapStageError(sourcePath, stageIndexAnalysis, "", fmt.Errorf("文件数量 %d 超出上限 %d", fileCount, maxFileCount))
	}

	expectedIndexEnd := uint64(reader.Size()) - uint64(bodyInfoLength)
	currentPos := uint64(reader.Size()) - uint64(reader.Len())
	if expectedIndexEnd < currentPos {
		return nil, wrapStageError(sourcePath, stageHeaderValidation, "", fmt.Errorf(
			"索引区结束位置异常: 当前位置 %d, 预期结束位置 %d",
			currentPos, expectedIndexEnd,
		))
	}

	fileNames := make([]string, 0, fileCount)
	plans := make([]plannedFile, 0, fileCount)
	usedFiles := make(map[string]struct{}, fileCount)
	usedDirs := make(map[string]struct{}, fileCount)

	for i := uint32(0); i < fileCount; i++ {
		var wxFile WxapkgFile
		if err := binary.Read(reader, binary.BigEndian, &wxFile.NameLen); err != nil {
			return nil, wrapStageError(sourcePath, stageIndexAnalysis, fmt.Sprintf("#%d", i), fmt.Errorf("读取文件名长度失败: %w", err))
		}

		if wxFile.NameLen == 0 || wxFile.NameLen > maxFileNameLength {
			return nil, wrapStageError(sourcePath, stageIndexAnalysis, fmt.Sprintf("#%d", i), fmt.Errorf(
				"文件名长度 %d 不合理，允许范围为 1-%d",
				wxFile.NameLen, maxFileNameLength,
			))
		}

		nameBytes := make([]byte, wxFile.NameLen)
		if _, err := io.ReadFull(reader, nameBytes); err != nil {
			return nil, wrapStageError(sourcePath, stageIndexAnalysis, fmt.Sprintf("#%d", i), fmt.Errorf("读取文件名失败: %w", err))
		}
		wxFile.Name = string(nameBytes)

		if err := binary.Read(reader, binary.BigEndian, &wxFile.Offset); err != nil {
			return nil, wrapStageError(sourcePath, stageIndexAnalysis, wxFile.Name, fmt.Errorf("读取文件偏移量失败: %w", err))
		}
		if err := binary.Read(reader, binary.BigEndian, &wxFile.Size); err != nil {
			return nil, wrapStageError(sourcePath, stageIndexAnalysis, wxFile.Name, fmt.Errorf("读取文件大小失败: %w", err))
		}
		if wxFile.Size > maxSingleFileSize {
			return nil, wrapStageError(sourcePath, stageIndexAnalysis, wxFile.Name, fmt.Errorf(
				"文件大小 %d 超出上限 %d",
				wxFile.Size, maxSingleFileSize,
			))
		}

		fileEnd := uint64(wxFile.Offset) + uint64(wxFile.Size)
		if fileEnd > uint64(len(data)) {
			return nil, wrapStageError(sourcePath, stageIndexAnalysis, wxFile.Name, fmt.Errorf(
				"文件结束位置 %d 超出文件总长度 %d",
				fileEnd, len(data),
			))
		}

		currentPos = uint64(reader.Size()) - uint64(reader.Len())
		if currentPos > expectedIndexEnd {
			return nil, wrapStageError(sourcePath, stageIndexAnalysis, wxFile.Name, fmt.Errorf(
				"索引读取超出预期范围: 当前位置 %d, 预期索引结束位置 %d",
				currentPos, expectedIndexEnd,
			))
		}

		relativePath, fullPath, err := planOutputPath(outputAbs, wxFile.Name, usedFiles, usedDirs)
		if err != nil {
			return nil, wrapStageError(sourcePath, stagePathPlanning, wxFile.Name, err)
		}

		fileNames = append(fileNames, wxFile.Name)
		plans = append(plans, plannedFile{
			Index:        int(i),
			EntryName:    wxFile.Name,
			RelativePath: relativePath,
			FullPath:     fullPath,
			Offset:       wxFile.Offset,
			Size:         wxFile.Size,
		})
	}

	currentPos = uint64(reader.Size()) - uint64(reader.Len())
	if currentPos != expectedIndexEnd {
		return nil, wrapStageError(sourcePath, stageIndexAnalysis, "", fmt.Errorf(
			"索引段长度不符: 读取到位置 %d, 预期结束位置 %d",
			currentPos, expectedIndexEnd,
		))
	}

	return &packagePlan{
		SourcePath: sourcePath,
		OutputDir:  outputAbs,
		FileNames:  fileNames,
		Files:      plans,
	}, nil
}

func planOutputPath(outputDir string, entryName string, usedFiles map[string]struct{}, usedDirs map[string]struct{}) (string, string, error) {
	normalized, err := normalizeEntryPath(entryName)
	if err != nil {
		return "", "", err
	}

	relativePath, err := allocatePlannedPath(normalized, usedFiles, usedDirs)
	if err != nil {
		return "", "", err
	}

	fullPath := filepath.Join(outputDir, filepath.FromSlash(relativePath))
	fullPathAbs, err := filepath.Abs(fullPath)
	if err != nil {
		return "", "", fmt.Errorf("解析目标文件绝对路径失败: %w", err)
	}

	withinBase, err := isWithinBaseDir(outputDir, fullPathAbs)
	if err != nil {
		return "", "", err
	}
	if !withinBase {
		return "", "", fmt.Errorf("目标路径 %s 超出输出目录 %s", fullPathAbs, outputDir)
	}

	return relativePath, fullPathAbs, nil
}

func normalizeEntryPath(entryName string) (string, error) {
	virtualPath := strings.TrimSpace(strings.ReplaceAll(entryName, "\\", "/"))
	virtualPath = strings.TrimLeft(virtualPath, "/")
	if virtualPath == "" {
		return "", fmt.Errorf("包内路径为空")
	}

	cleanPath := path.Clean(virtualPath)
	if cleanPath == "." || cleanPath == "" {
		return "", fmt.Errorf("包内路径为空")
	}
	if cleanPath == ".." || strings.HasPrefix(cleanPath, "../") {
		return "", fmt.Errorf("包内路径 %q 存在目录穿越", entryName)
	}

	return cleanPath, nil
}

func allocatePlannedPath(relativePath string, usedFiles map[string]struct{}, usedDirs map[string]struct{}) (string, error) {
	for attempt := 0; ; attempt++ {
		candidate := relativePath
		if attempt > 0 {
			candidate = addNumericSuffix(relativePath, attempt)
		}

		if err := ensurePathAvailable(candidate, usedFiles, usedDirs); err != nil {
			if errors.Is(err, errDuplicatePlannedPath) {
				continue
			}
			return "", err
		}

		registerPlannedPath(candidate, usedFiles, usedDirs)
		return candidate, nil
	}
}

func ensurePathAvailable(relativePath string, usedFiles map[string]struct{}, usedDirs map[string]struct{}) error {
	if _, exists := usedFiles[relativePath]; exists {
		return errDuplicatePlannedPath
	}
	if _, exists := usedDirs[relativePath]; exists {
		return fmt.Errorf("目标路径 %s 与已有目录冲突", relativePath)
	}

	dir := path.Dir(relativePath)
	for dir != "." && dir != "/" {
		if _, exists := usedFiles[dir]; exists {
			return fmt.Errorf("目标路径 %s 的父路径 %s 已作为文件输出", relativePath, dir)
		}
		dir = path.Dir(dir)
	}

	return nil
}

func registerPlannedPath(relativePath string, usedFiles map[string]struct{}, usedDirs map[string]struct{}) {
	usedFiles[relativePath] = struct{}{}

	dir := path.Dir(relativePath)
	for dir != "." && dir != "/" {
		usedDirs[dir] = struct{}{}
		dir = path.Dir(dir)
	}
}

func addNumericSuffix(relativePath string, index int) string {
	dir, file := path.Split(relativePath)
	ext := path.Ext(file)
	base := strings.TrimSuffix(file, ext)
	return path.Join(dir, fmt.Sprintf("%s-%d%s", base, index, ext))
}

func isWithinBaseDir(baseDir string, targetPath string) (bool, error) {
	rel, err := filepath.Rel(baseDir, targetPath)
	if err != nil {
		return false, fmt.Errorf("计算目标相对路径失败: %w", err)
	}

	if rel == ".." {
		return false, nil
	}

	return !strings.HasPrefix(rel, ".."+string(os.PathSeparator)), nil
}

func writePlannedFiles(plan *packagePlan, reader io.ReaderAt) error {
	workerCount := runtime.NumCPU() * 2
	if workerCount < 4 {
		workerCount = 4
	}
	if workerCount > 32 {
		workerCount = 32
	}

	fileChan := make(chan plannedFile, workerCount)
	errChan := make(chan error, len(plan.Files))

	var wg sync.WaitGroup
	var bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				if err := processPlannedFile(plan.SourcePath, file, reader, &bufferPool); err != nil {
					errChan <- err
				}
			}
		}()
	}

	for _, file := range plan.Files {
		fileChan <- file
	}
	close(fileChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}

	return nil
}

func processPlannedFile(sourcePath string, file plannedFile, reader io.ReaderAt, bufferPool *sync.Pool) error {
	dir := filepath.Dir(file.FullPath)
	if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
		return wrapStageError(sourcePath, stageFileWrite, file.RelativePath, fmt.Errorf("创建目录失败: %w", err))
	}

	sectionReader := io.NewSectionReader(reader, int64(file.Offset), int64(file.Size))

	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf)
	buf.Reset()

	if _, err := io.Copy(buf, sectionReader); err != nil {
		return wrapStageError(sourcePath, stageFileRead, file.RelativePath, fmt.Errorf("读取文件内容失败: %w", err))
	}
	content := buf.Bytes()

	ext := filepath.Ext(file.EntryName)
	var jsResult *formatter2.DeobfuscationResult
	formatter, err := formatter2.GetFormatter(ext)
	if err == nil {
		if fileFormatter, ok := formatter.(formatter2.FileFormatter); ok {
			content, jsResult, err = fileFormatter.FormatFile(content, file.RelativePath)
		} else {
			content, err = formatter.Format(content)
		}
		if err != nil {
			return wrapStageError(sourcePath, stageFileFormat, file.RelativePath, fmt.Errorf("格式化文件失败: %w", err))
		}
	}

	outputFile, err := os.OpenFile(file.FullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return wrapStageError(sourcePath, stageFileWrite, file.RelativePath, fmt.Errorf("创建文件失败: %w", err))
	}
	defer outputFile.Close()

	writer := bufio.NewWriterSize(outputFile, 256*1024)
	if _, err := writer.Write(content); err != nil {
		return wrapStageError(sourcePath, stageFileWrite, file.RelativePath, fmt.Errorf("写入文件失败: %w", err))
	}
	if err := writer.Flush(); err != nil {
		return wrapStageError(sourcePath, stageFileWrite, file.RelativePath, fmt.Errorf("刷新缓冲区失败: %w", err))
	}

	configManager := config.NewSharedConfigManager()
	shouldScan := false
	if sensitive, ok := configManager.Get("sensitive"); ok {
		if enabled, ok := sensitive.(bool); ok && enabled {
			shouldScan = true
		}
	}
	if postman, ok := configManager.Get("postman"); ok {
		if enabled, ok := postman.(bool); ok && enabled {
			shouldScan = true
		}
	}

	if shouldScan {
		collector := key.GetCollector()
		if collector != nil {
			if jsResult != nil && jsResult.IsObfuscated {
				collector.AddObfuscatedFile(scanner.ObfuscatedFile{
					FilePath:   file.RelativePath,
					Score:      jsResult.Score,
					Techniques: jsResult.Techniques,
					Status:     jsResult.Status,
					Tag:        formatter2.BuildObfuscatedTag(jsResult),
				})
			}
			if err := scanner.ScanFile(file.RelativePath, content, collector); err != nil {
				fmt.Printf("警告: %v\n", wrapStageError(sourcePath, stageSensitiveScan, file.RelativePath, err))
			}
		}
	}

	return nil
}

func wrapStageError(sourcePath, stage, file string, err error) error {
	if err == nil {
		return nil
	}

	var stageErr *packageStageError
	if errors.As(err, &stageErr) {
		return err
	}

	return &packageStageError{
		SourcePath: sourcePath,
		Stage:      stage,
		File:       file,
		Err:        err,
	}
}
