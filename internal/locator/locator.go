package locator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/25smoking/Gwxapkg/internal/decrypt"
)

// MiniProgramInfo 存储小程序的基本信息
type MiniProgramInfo struct {
	AppID      string
	AppName    string
	Version    string
	UpdateTime time.Time
	Path       string
	Files      []string
}

// ScanOptions 控制扫描行为。
type ScanOptions struct {
	Verbose bool
}

// ScanDiagnostic 表示扫描候选路径与命中情况。
type ScanDiagnostic struct {
	Path   string
	Status string
	Detail string
}

// ScanReport 包含扫描结果及可选诊断信息。
type ScanReport struct {
	Programs    []MiniProgramInfo
	Diagnostics []ScanDiagnostic
}

type basePathCandidate struct {
	Path   string
	Source string
}

var genericAppTitles = map[string]struct{}{
	"首页":     {},
	"加载中":    {},
	"加载中...": {},
	"支付":     {},
	"微信支付":   {},
	"购物车":    {},
	"提示":     {},
}

var genericNavigationTitles = map[string]struct{}{
	"首页":     {},
	"我的":     {},
	"个人中心":   {},
	"用户中心":   {},
	"加载中":    {},
	"加载中...": {},
	"支付":     {},
	"微信支付":   {},
	"购物车":    {},
	"提示":     {},
	"登录":     {},
	"会员登录":   {},
	"搜索":     {},
	"订单":     {},
	"订单列表":   {},
	"订单详情":   {},
	"我的订单":   {},
	"收货地址":   {},
	"新增收货地址": {},
	"选择门店":   {},
	"选择城市":   {},
	"点餐":     {},
	"手机点餐":   {},
	"下单页":    {},
	"优惠券页":   {},
	"会员中心":   {},
	"测试结果":   {},
	"测试详情":   {},
	"测试历史":   {},
	"更多结果":   {},
}

var machineLikeNamePattern = regexp.MustCompile(`^[a-z0-9_-]{4,}$`)

var preciseCodeNamePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?is)appId\s*:\s*["'][^"']+["']\s*,\s*appName\s*:\s*["']([^"'\\\r\n]{1,80})["']\s*,\s*appVersion`),
	regexp.MustCompile(`(?is)appName\s*:\s*["']([^"'\\\r\n]{1,80})["']\s*,\s*appVersion\s*:\s*["'][^"']+["']\s*,\s*appVersionCode`),
	regexp.MustCompile(`(?is)\bauthFloatInfo\s*:\s*\{.*?\bname\s*:\s*["']([^"'\\\r\n]{1,80})["']`),
	regexp.MustCompile(`(?m)\bAPPNAME\s*:\s*["']([^"'\\\r\n]{1,80})["']`),
	regexp.MustCompile(`(?m)\bAPPNAME\s*=\s*["']([^"'\\\r\n]{1,80})["']`),
}

var preferredSourceNameFiles = []string{
	"common/vendor.js",
	"app-service.js",
	"app.js",
	"common/main.js",
	"main.js",
	"vendor.js",
	"manifest.js",
}

const runtimeAssignedShopNamePlaceholder = "点餐模板(商户名运行时下发)"

var genericCodeNames = map[string]struct{}{
	"Netscape": {},
	"Mozilla":  {},
}

var preciseAppNameCache sync.Map

var userHomeDirFunc = os.UserHomeDir
var scanBasePathCollector = collectPlatformBasePaths

type meituanAppInfoResponse struct {
	Code int `json:"code"`
	Data struct {
		Nickname string `json:"nickname"`
	} `json:"data"`
}

func queryWeChatMiniProgramName(appID string) string {
	if appID == "" {
		return ""
	}

	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "https://mp.weixin.qq.com/wxawap/waverifyinfo?action=get&appid="+appID, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.ContentLength == 0 {
		return ""
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	text := string(data)
	for _, pattern := range []*regexp.Regexp{
		regexp.MustCompile(`profile_nickname[^>]*>\s*([^<]+)\s*<`),
		regexp.MustCompile(`weui-media-box__title[^>]*>\s*([^<]+)\s*<`),
		regexp.MustCompile(`<title>\s*([^<]+)\s*</title>`),
	} {
		match := pattern.FindStringSubmatch(text)
		if len(match) < 2 {
			continue
		}
		candidate := sanitizeDisplayName(match[1])
		if isLikelyHumanReadableName(candidate) && isMeaningfulNavigationTitle(candidate) {
			return candidate
		}
	}

	return ""
}

func sanitizeDisplayName(name string) string {
	name = strings.TrimSpace(strings.Trim(name, `"'`))
	if name == "" {
		return ""
	}
	return name
}

func isMeaningfulTitle(name string) bool {
	name = sanitizeDisplayName(name)
	if name == "" {
		return false
	}
	_, exists := genericAppTitles[name]
	return !exists
}

func isMeaningfulNavigationTitle(name string) bool {
	name = sanitizeDisplayName(name)
	if name == "" {
		return false
	}
	if !isMeaningfulTitle(name) {
		return false
	}
	_, exists := genericNavigationTitles[name]
	return !exists
}

func isLikelyHumanReadableName(name string) bool {
	name = sanitizeDisplayName(name)
	if name == "" {
		return false
	}
	return !machineLikeNamePattern.MatchString(name)
}

func isLikelyPreciseCodeName(name string) bool {
	name = sanitizeDisplayName(name)
	if !isLikelyHumanReadableName(name) {
		return false
	}
	if _, exists := genericCodeNames[name]; exists {
		return false
	}
	if len([]rune(name)) >= 4 {
		return true
	}
	for _, r := range name {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

func extractStringField(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		value, ok := data[key]
		if !ok {
			continue
		}
		text, ok := value.(string)
		if !ok {
			continue
		}
		text = sanitizeDisplayName(text)
		if text != "" {
			return text
		}
	}
	return ""
}

func extractMarkerValue(data, marker string) string {
	idx := strings.Index(data, marker)
	if idx == -1 {
		return ""
	}
	start := idx + len(marker)
	end := strings.IndexByte(data[start:], '"')
	if end == -1 {
		return ""
	}
	return sanitizeDisplayName(data[start : start+end])
}

func hasNonEmptyStringField(data map[string]interface{}, keys ...string) bool {
	for _, key := range keys {
		text := extractStringField(data, key)
		if text != "" {
			return true
		}
	}
	return false
}

func hasRuntimeAssignedShopName(data map[string]interface{}) bool {
	if hasNonEmptyStringField(data, "restaurantViewId", "tenantId", "bizId") {
		return true
	}
	if global, ok := data["global"].(map[string]interface{}); ok {
		if hasNonEmptyStringField(global, "restaurantViewId", "tenantId", "bizId") {
			return true
		}
	}
	return false
}

func hasRuntimeAssignedShopNameText(data string) bool {
	return strings.Contains(data, `"restaurantViewId"`) && strings.Contains(data, `"tenantId"`)
}

func extractNameFromJSONContent(data []byte) string {
	var meta map[string]interface{}
	if err := json.Unmarshal(data, &meta); err != nil {
		return ""
	}

	return extractStringField(meta,
		"nickname",
		"appName",
		"name",
		"title",
		"brandName",
		"storeName",
		"mallName",
		"miniProgramName",
	)
}

func tryReadJSONNames(dir string, files []string) string {
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		if extracted := extractNameFromJSONContent(data); extracted != "" {
			return extracted
		}
	}
	return ""
}

func tryReadLocalAppName(appPath, appID string) string {
	localPath := filepath.Join(filepath.Dir(filepath.Dir(appPath)), "local", appID)

	info, err := os.Stat(localPath)
	if err != nil || !info.IsDir() {
		return ""
	}

	return tryReadJSONNames(localPath, []string{
		"appinfo.json",
		"appInfo.json",
		"app-info.json",
		"info.json",
		"adapter-config.json",
		"initial-rendering-cache-config.json",
	})
}

func extractNameFromAppConfig(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	dataStr := string(data)

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err == nil {
		if direct := extractStringField(config,
			"appName",
			"nickname",
			"name",
			"title",
			"brandName",
			"storeName",
			"mallName",
			"miniProgramName",
		); direct != "" {
			return direct
		}

		if global, ok := config["global"].(map[string]interface{}); ok {
			if title := extractStringField(global,
				"appName",
				"nickname",
				"name",
				"title",
				"brandName",
				"storeName",
				"mallName",
			); title != "" {
				return title
			}

			if window, ok := global["window"].(map[string]interface{}); ok {
				if title := extractStringField(window, "navigationBarTitleText", "title"); isMeaningfulNavigationTitle(title) {
					return title
				}
			}
		}

		if hasRuntimeAssignedShopName(config) {
			return runtimeAssignedShopNamePlaceholder
		}
	}

	for _, marker := range []string{
		`"appName":"`,
		`"nickname":"`,
		`"miniProgramName":"`,
		`"brandName":"`,
		`"storeName":"`,
		`"mallName":"`,
	} {
		candidate := extractMarkerValue(dataStr, marker)
		if candidate != "" {
			return candidate
		}
	}

	if hasRuntimeAssignedShopNameText(dataStr) {
		return runtimeAssignedShopNamePlaceholder
	}

	return ""
}

func extractNameFromPluginConfig(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var plugin map[string]interface{}
	if err := json.Unmarshal(data, &plugin); err == nil {
		if direct := extractStringField(plugin,
			"name",
			"pluginName",
			"title",
			"nickname",
			"componentName",
			"displayName",
		); direct != "" {
			return direct
		}

		if publicComponents, ok := plugin["publicComponents"].(map[string]interface{}); ok {
			aliases := make([]string, 0, len(publicComponents))
			for key := range publicComponents {
				key = sanitizeDisplayName(key)
				if key != "" {
					aliases = append(aliases, key)
				}
			}
			sort.Strings(aliases)
			if len(aliases) > 0 {
				return fmt.Sprintf("插件(%s)", aliases[0])
			}
		}
	}

	text := string(data)
	if strings.Contains(text, "geetest") {
		return "插件(geetest)"
	}

	return ""
}

func fallbackDisplayName(files []string) string {
	if len(files) == 0 {
		return ""
	}

	onlyPlugin := true
	for _, file := range files {
		if !strings.EqualFold(filepath.Base(file), "__PLUGINCODE__.wxapkg") {
			onlyPlugin = false
			break
		}
	}

	if onlyPlugin {
		return "插件包(未发现名称元数据)"
	}

	return "未发现名称元数据"
}

func extractNameFromCodeContent(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	text := string(data)
	for _, pattern := range preciseCodeNamePatterns {
		matches := pattern.FindAllStringSubmatch(text, 3)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			candidate := sanitizeDisplayName(match[1])
			if !isLikelyPreciseCodeName(candidate) {
				continue
			}
			if isMeaningfulNavigationTitle(candidate) {
				return candidate
			}
		}
	}

	return ""
}

func queryMeituanMiniProgramName(appID string) string {
	if appID == "" {
		return ""
	}

	if cached, ok := preciseAppNameCache.Load(appID); ok {
		name, _ := cached.(string)
		return name
	}

	payload := strings.NewReader(fmt.Sprintf(`{"appId":"%s","restType":1}`, appID))
	req, err := http.NewRequest(http.MethodPost, "https://pos.meituan.com/api/v1/crm/mini/app/tenant/query", payload)
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		preciseAppNameCache.Store(appID, "")
		return ""
	}
	defer resp.Body.Close()

	var result meituanAppInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		preciseAppNameCache.Store(appID, "")
		return ""
	}

	if result.Code != 0 {
		preciseAppNameCache.Store(appID, "")
		return ""
	}

	name := sanitizeDisplayName(result.Data.Nickname)
	preciseAppNameCache.Store(appID, name)
	return name
}

func lookupRemoteAppName(appID string) string {
	if appID == "" {
		return ""
	}

	if cached, ok := preciseAppNameCache.Load(appID); ok {
		name, _ := cached.(string)
		return name
	}

	for _, resolver := range []func(string) string{
		queryWeChatMiniProgramName,
		queryMeituanMiniProgramName,
	} {
		if name := sanitizeDisplayName(resolver(appID)); name != "" {
			preciseAppNameCache.Store(appID, name)
			return name
		}
	}

	preciseAppNameCache.Store(appID, "")
	return ""
}

func resolvePreciseAppName(appID, name string) string {
	if name != runtimeAssignedShopNamePlaceholder {
		return name
	}

	if precise := queryMeituanMiniProgramName(appID); precise != "" {
		return precise
	}

	return name
}

func sourceNameFilePriority(path string) int {
	path = strings.TrimPrefix(filepath.ToSlash(path), "/")
	for idx, candidate := range preferredSourceNameFiles {
		if path == candidate {
			return idx
		}
	}
	return -1
}

func wxapkgNamePriority(path string) int {
	base := strings.ToUpper(filepath.Base(path))
	switch base {
	case "__APP__.WXAPKG":
		return 0
	case "__FULL__.WXAPKG":
		return 1
	case "__SUBPACKAGE__.WXAPKG":
		return 2
	case "__PLUGINCODE__.WXAPKG":
		return 4
	default:
		return 3
	}
}

func tryReadSourceAppName(appPath string) string {
	if data, err := os.ReadFile(filepath.Join(appPath, "app-config.json")); err == nil {
		if name := extractNameFromAppConfig(data); name != "" {
			return name
		}
	}

	if name := tryReadJSONNames(appPath, []string{
		"manifest.json",
		"ext.json",
		"extconfig.json",
		"project.private.config.json",
	}); name != "" {
		return name
	}

	for _, relPath := range preferredSourceNameFiles {
		data, err := os.ReadFile(filepath.Join(appPath, filepath.FromSlash(relPath)))
		if err != nil {
			continue
		}
		if name := extractNameFromCodeContent(data); name != "" {
			return name
		}
	}

	return ""
}

// tryReadAppName 尝试读取小程序应用名称
func tryReadAppName(appPath, appID string, files []string) string {
	if name := sanitizeDisplayName(tryReadLocalAppName(appPath, appID)); name != "" {
		return name
	}

	sortedFiles := append([]string(nil), files...)
	sort.SliceStable(sortedFiles, func(i, j int) bool {
		left := wxapkgNamePriority(sortedFiles[i])
		right := wxapkgNamePriority(sortedFiles[j])
		if left != right {
			return left < right
		}
		return sortedFiles[i] < sortedFiles[j]
	})

	needsRuntimeLookup := false
	for _, file := range sortedFiles {
		name := sanitizeDisplayName(extractNameFromWxapkg(file, appID))
		if name == "" {
			continue
		}
		if name == runtimeAssignedShopNamePlaceholder {
			needsRuntimeLookup = true
			continue
		}
		return name
	}

	if needsRuntimeLookup {
		if name := lookupRemoteAppName(appID); name != "" {
			return name
		}
		return ""
	}

	return lookupRemoteAppName(appID)
}

// extractNameFromWxapkg 尝试在内存中快速解密并提取包内应用名
func extractNameFromWxapkg(file, appID string) string {
	dec, err := decrypt.DecryptWxapkg(file, appID)
	if err != nil || len(dec) < 14 {
		return ""
	}

	r := bytes.NewReader(dec)
	var firstMark byte
	firstMark, _ = r.ReadByte()
	if firstMark != 0xBE {
		return ""
	}

	// 跳转到文件数量
	r.Seek(14, 0)
	var buf [4]byte
	if _, err := r.Read(buf[:]); err != nil {
		return ""
	}
	fileCount := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])

	var appConfigData []byte
	var pluginConfigData []byte
	bestCodeName := ""
	bestCodePriority := len(preferredSourceNameFiles) + 1

	for i := 0; i < int(fileCount); i++ {
		if _, err := r.Read(buf[:]); err != nil {
			return ""
		}
		nameLen := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])

		nameBuf := make([]byte, nameLen)
		if _, err := r.Read(nameBuf); err != nil {
			return ""
		}

		if _, err := r.Read(buf[:]); err != nil {
			return ""
		}
		offset := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])

		if _, err := r.Read(buf[:]); err != nil {
			return ""
		}
		size := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])

		name := string(nameBuf)
		priority := sourceNameFilePriority(name)
		if name == "/app-config.json" || name == "/plugin.json" || priority >= 0 {
			fileData := make([]byte, size)
			currentPos, _ := r.Seek(0, 1)
			if _, err := r.Seek(int64(offset), 0); err != nil {
				return ""
			}
			if _, err := r.Read(fileData); err != nil {
				return ""
			}
			if _, err := r.Seek(currentPos, 0); err != nil {
				return ""
			}

			if name == "/app-config.json" {
				appConfigData = fileData
			} else if name == "/plugin.json" {
				pluginConfigData = fileData
			} else if priority < bestCodePriority {
				if codeName := extractNameFromCodeContent(fileData); codeName != "" {
					bestCodeName = codeName
					bestCodePriority = priority
				}
			}
		}
	}

	if name := extractNameFromAppConfig(appConfigData); name != "" {
		return name
	}

	if name := extractNameFromPluginConfig(pluginConfigData); name != "" {
		return name
	}

	if bestCodeName != "" {
		return bestCodeName
	}

	// 部分插件包没有显式名称，只能根据源码中的特征给出轻量提示。
	if len(pluginConfigData) > 0 && strings.Contains(string(dec), "geetest") {
		return "插件(geetest)"
	}

	return ""
}

// Scan 扫描所有可能的微信小程序目录。
func Scan() ([]MiniProgramInfo, error) {
	report, err := ScanWithOptions(ScanOptions{})
	if err != nil {
		return nil, err
	}

	return report.Programs, nil
}

// ScanWithOptions 按指定选项执行扫描，并返回可选诊断信息。
func ScanWithOptions(opts ScanOptions) (*ScanReport, error) {
	homeDir, err := userHomeDirFunc()
	if err != nil {
		return nil, fmt.Errorf("获取用户目录失败: %w", err)
	}

	return scanWithBasePathCollector(homeDir, opts, scanBasePathCollector)
}

func scanWithBasePathCollector(homeDir string, opts ScanOptions, collector func(string) ([]basePathCandidate, []ScanDiagnostic, error)) (*ScanReport, error) {
	candidates, diagnostics, err := collector(homeDir)
	if err != nil {
		return nil, err
	}

	report := &ScanReport{}
	if opts.Verbose {
		report.Diagnostics = append(report.Diagnostics, diagnostics...)
	}

	seen := make(map[string]struct{}, len(candidates))
	checkedCount := 0

	for _, candidate := range candidates {
		cleanPath := filepath.Clean(candidate.Path)
		if cleanPath == "." || cleanPath == "" {
			if opts.Verbose {
				report.Diagnostics = append(report.Diagnostics, ScanDiagnostic{
					Path:   candidate.Path,
					Status: "invalid",
					Detail: "候选路径为空，已跳过",
				})
			}
			continue
		}

		if _, ok := seen[cleanPath]; ok {
			if opts.Verbose {
				report.Diagnostics = append(report.Diagnostics, ScanDiagnostic{
					Path:   cleanPath,
					Status: "duplicate",
					Detail: "候选路径重复，已去重",
				})
			}
			continue
		}
		seen[cleanPath] = struct{}{}
		checkedCount++

		info, statErr := os.Stat(cleanPath)
		if statErr != nil {
			if opts.Verbose {
				report.Diagnostics = append(report.Diagnostics, buildPathStatDiagnostic(cleanPath, statErr)...)
			}
			continue
		}
		if !info.IsDir() {
			if opts.Verbose {
				report.Diagnostics = append(report.Diagnostics, ScanDiagnostic{
					Path:   cleanPath,
					Status: "not-dir",
					Detail: "候选路径存在，但不是目录",
				})
			}
			continue
		}

		programs, appCount, scanErr := scanDirectory(cleanPath)
		if scanErr != nil {
			if opts.Verbose {
				report.Diagnostics = append(report.Diagnostics, ScanDiagnostic{
					Path:   cleanPath,
					Status: "scan-error",
					Detail: fmt.Sprintf("扫描目录失败: %v", scanErr),
				})
			}
			continue
		}

		report.Programs = append(report.Programs, programs...)

		if opts.Verbose {
			status := "empty"
			detail := "目录可访问，但未发现 wxapkg 包"
			if len(programs) > 0 {
				status = "hit"
				detail = fmt.Sprintf("发现 %d 个 AppID，%d 个版本目录", appCount, len(programs))
			}
			report.Diagnostics = append(report.Diagnostics, ScanDiagnostic{
				Path:   cleanPath,
				Status: status,
				Detail: detail,
			})
		}
	}

	sort.Slice(report.Programs, func(i, j int) bool {
		return report.Programs[i].UpdateTime.After(report.Programs[j].UpdateTime)
	})

	if opts.Verbose {
		report.Diagnostics = append(report.Diagnostics, ScanDiagnostic{
			Status: "summary",
			Detail: fmt.Sprintf("共检查 %d 个候选目录，命中 %d 个小程序版本", checkedCount, len(report.Programs)),
		})
	}

	return report, nil
}

func collectPlatformBasePaths(homeDir string) ([]basePathCandidate, []ScanDiagnostic, error) {
	switch runtime.GOOS {
	case "darwin":
		return collectDarwinBasePathCandidates(homeDir)
	case "windows":
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, []ScanDiagnostic{{
				Status: "config-error",
				Detail: fmt.Sprintf("获取 Windows 配置目录失败: %v", err),
			}}, nil
		}
		return collectWindowsBasePathCandidates(homeDir, configDir)
	default:
		return collectOtherBasePathCandidates(homeDir)
	}
}

func buildPathStatDiagnostic(path string, err error) []ScanDiagnostic {
	switch {
	case os.IsNotExist(err):
		return []ScanDiagnostic{{
			Path:   path,
			Status: "missing",
			Detail: "候选路径不存在",
		}}
	case os.IsPermission(err):
		return []ScanDiagnostic{{
			Path:   path,
			Status: "no-access",
			Detail: "候选路径无访问权限",
		}}
	default:
		return []ScanDiagnostic{{
			Path:   path,
			Status: "stat-error",
			Detail: fmt.Sprintf("检查候选路径失败: %v", err),
		}}
	}
}

func scanDirectory(basePath string) ([]MiniProgramInfo, int, error) {
	// 结构: base_path/{AppID}/{Version}/__APP__.wxapkg

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, 0, err
	}

	results := make([]MiniProgramInfo, 0)
	appCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		appID := entry.Name()
		// 忽略非 AppID 目录
		if !strings.HasPrefix(appID, "wx") {
			continue
		}

		appPath := filepath.Join(basePath, appID)
		verEntries, err := os.ReadDir(appPath)
		if err != nil {
			continue
		}

		appHasPackage := false

		for _, verEntry := range verEntries {
			if !verEntry.IsDir() {
				continue
			}

			version := verEntry.Name()
			verPath := filepath.Join(appPath, version)

			var wxapkgFiles []string
			var latestTime time.Time

			err := filepath.WalkDir(verPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(d.Name(), ".wxapkg") {
					wxapkgFiles = append(wxapkgFiles, path)

					info, err := d.Info()
					if err == nil {
						if info.ModTime().After(latestTime) {
							latestTime = info.ModTime()
						}
					}
				}
				return nil
			})

			if err != nil {
				continue
			}

			if len(wxapkgFiles) > 0 {
				appHasPackage = true
				results = append(results, MiniProgramInfo{
					AppID:      appID,
					AppName:    tryReadAppName(appPath, appID, wxapkgFiles),
					Version:    version,
					UpdateTime: latestTime,
					Path:       verPath,
					Files:      wxapkgFiles,
				})
			}
		}

		if appHasPackage {
			appCount++
		}
	}

	return results, appCount, nil
}
