package locator

import (
	"fmt"
	"path/filepath"
)

func collectWindowsBasePathCandidates(homeDir, configDir string) ([]basePathCandidate, []ScanDiagnostic, error) {
	candidates := make([]basePathCandidate, 0, 5)
	diagnostics := make([]ScanDiagnostic, 0, 8)

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

	if configDir != "" {
		addFixed(
			filepath.Join(configDir, "Tencent/xwechat/radium/Applet/packages"),
			"Windows xwechat 固定路径",
		)

		patterns := []struct {
			pattern string
			detail  string
		}{
			{
				pattern: filepath.Join(configDir, "Tencent/xwechat/radium/users/*/applet/packages"),
				detail:  "Windows xwechat 用户隔离目录",
			},
			{
				pattern: filepath.Join(configDir, "Tencent/WeChat/radium/Applet/*/packages"),
				detail:  "Windows WeChat 旧版隔离目录",
			},
		}

		for _, item := range patterns {
			matches, err := filepath.Glob(item.pattern)
			if err != nil {
				diagnostics = append(diagnostics, ScanDiagnostic{
					Path:   item.pattern,
					Status: "glob-error",
					Detail: fmt.Sprintf("%s展开失败: %v", item.detail, err),
				})
				continue
			}

			diagnostics = append(diagnostics, ScanDiagnostic{
				Path:   item.pattern,
				Status: "glob",
				Detail: fmt.Sprintf("%s展开得到 %d 个候选路径", item.detail, len(matches)),
			})

			for _, match := range matches {
				candidates = append(candidates, basePathCandidate{
					Path:   match,
					Source: item.detail,
				})
				diagnostics = append(diagnostics, ScanDiagnostic{
					Path:   match,
					Status: "candidate",
					Detail: fmt.Sprintf("来自 %s 展开结果", item.detail),
				})
			}
		}
	}

	addFixed(
		filepath.Join(homeDir, "Documents/WeChat Files/Applet"),
		"Windows Documents 固定路径",
	)

	return candidates, diagnostics, nil
}
