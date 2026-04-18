package locator

func collectOtherBasePathCandidates(homeDir string) ([]basePathCandidate, []ScanDiagnostic, error) {
	return nil, []ScanDiagnostic{{
		Status: "unsupported",
		Detail: "当前平台未内置微信缓存扫描候选路径",
	}}, nil
}
