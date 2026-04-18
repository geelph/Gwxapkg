package locator

import (
	"fmt"
	"path/filepath"
)

func collectDarwinBasePathCandidates(homeDir string) ([]basePathCandidate, []ScanDiagnostic, error) {
	candidates := make([]basePathCandidate, 0, 4)
	diagnostics := make([]ScanDiagnostic, 0, 6)

	addFixed := func(path string, detail string) {
		candidates = append(candidates, basePathCandidate{
			Path:   path,
			Source: detail,
		})
		diagnostics = append(diagnostics, ScanDiagnostic{
			Path:   path,
			Status: "candidate",
			Detail: detail,
		})
	}

	addFixed(
		filepath.Join(homeDir, "Library/Containers/com.tencent.xinWeChat/Data/Documents/app_data/radium/Applet/packages"),
		"macOS 沙盒旧版固定路径",
	)
	addFixed(
		filepath.Join(homeDir, "Library/Application Support/WeChat/Applet/packages"),
		"macOS 非沙盒固定路径",
	)

	pattern := filepath.Join(homeDir, "Library/Containers/com.tencent.xinWeChat/Data/Documents/app_data/radium/users/*/applet/packages")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		diagnostics = append(diagnostics, ScanDiagnostic{
			Path:   pattern,
			Status: "glob-error",
			Detail: fmt.Sprintf("展开用户隔离目录失败: %v", err),
		})
		return candidates, diagnostics, nil
	}

	diagnostics = append(diagnostics, ScanDiagnostic{
		Path:   pattern,
		Status: "glob",
		Detail: fmt.Sprintf("用户隔离目录展开得到 %d 个候选路径", len(matches)),
	})

	for _, match := range matches {
		candidates = append(candidates, basePathCandidate{
			Path:   match,
			Source: "macOS 用户隔离目录",
		})
		diagnostics = append(diagnostics, ScanDiagnostic{
			Path:   match,
			Status: "candidate",
			Detail: "来自 macOS 用户隔离目录展开结果",
		})
	}

	return candidates, diagnostics, nil
}
