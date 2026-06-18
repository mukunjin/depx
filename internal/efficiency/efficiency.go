package efficiency

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// EfficiencyResult 效率分析结果
type EfficiencyResult struct {
	Package          string   // 包名
	UsedExports      []string // 使用的导出
	EstimatedExports int      // 估计的总导出数
	Efficiency       float64  // 效率百分比
}

// AnalyzeEfficiency 分析依赖使用效率
func AnalyzeEfficiency(dir string, deps []string) (map[string]*EfficiencyResult, error) {
	results := make(map[string]*EfficiencyResult)

	for _, dep := range deps {
		results[dep] = &EfficiencyResult{
			Package:     dep,
			UsedExports: []string{},
		}
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(info.Name())
		if ext != ".js" && ext != ".ts" && ext != ".jsx" && ext != ".tsx" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		analyzeFile(string(data), results)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 计算效率
	for _, result := range results {
		if len(result.UsedExports) == 0 {
			continue
		}

		// 去重
		unique := make(map[string]bool)
		for _, exp := range result.UsedExports {
			unique[exp] = true
		}
		result.UsedExports = make([]string, 0, len(unique))
		for exp := range unique {
			result.UsedExports = append(result.UsedExports, exp)
		}
		sort.Strings(result.UsedExports)

		result.EstimatedExports = estimateTotalExports(len(result.UsedExports))
		result.Efficiency = float64(len(result.UsedExports)) / float64(result.EstimatedExports) * 100
	}

	return results, nil
}

// analyzeFile 分析单个文件中的导入，将使用的导出记录到对应包的结果中
func analyzeFile(content string, results map[string]*EfficiencyResult) {
	infos := extractJSImportInfos(content)
	for _, info := range infos {
		if result, ok := results[info.Package]; ok {
			result.UsedExports = append(result.UsedExports, info.Exports...)
		}
	}
}

// estimateTotalExports 基于使用的导出数估计库的总导出数
func estimateTotalExports(usedCount int) int {
	if usedCount <= 2 {
		return 30
	} else if usedCount <= 5 {
		return 50
	} else if usedCount <= 10 {
		return 80
	} else if usedCount <= 20 {
		return 120
	}
	return 200
}

// AnalyzeExports 分析源文件中的导出
func AnalyzeExports(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return extractJSExports(string(data)), nil
}
