package manifest

// Manifest 定义包管理器清单文件的接口
type Manifest interface {
	// Type 返回包管理器类型 ("npm", "go", "cargo", "pip")
	Type() string
	// Dependencies 返回声明的运行时依赖包名列表
	Dependencies() ([]string, error)
	// DevDependencies 返回声明的开发时依赖包名列表（如果支持）
	DevDependencies() ([]string, error)
}

// UsageResult 表示依赖的使用情况
type UsageResult struct {
	Package  string   // 包名
	Used     bool     // 是否被使用
	UsedIn   []string // 使用该依赖的文件路径
	RefCount int      // 引用次数
}
