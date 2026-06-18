# depx

Dependency Efficiency Analyzer - 依赖效率分析工具

检测项目中声明但未使用的依赖，帮助开发者识别依赖浪费。

## 功能

- 扫描项目依赖声明文件
- 分析源码中的实际使用情况
- 输出未使用的依赖列表

## 支持范围

| 包管理器 | 清单文件 | 源文件 |
|----------|----------|--------|
| npm | package.json | .js, .ts, .jsx, .tsx, .mjs, .cjs, .vue, .svelte |
| Go | go.mod | .go |

## 安装

### 从 Release 下载

1. 前往 [Releases](https://github.com/depx/depx/releases) 下载对应平台的可执行文件
2. 将可执行文件放到系统 PATH 中的任一目录，即可在命令行直接使用 `depx` 命令

查看当前 PATH：

```bash
# Windows
echo %PATH%

# Linux / macOS
echo $PATH
```

若不想修改 PATH，也可使用完整路径运行：

```bash
# Windows
C:\Users\xxx\Downloads\depx.exe scan

# Linux / macOS
~/Downloads/depx scan
```

### 从源码构建

```bash
git clone https://github.com/depx/depx.git
cd depx
go build -ldflags="-s -w" -o depx .
```

### 通过 go install 安装

```bash
go install github.com/depx/depx@latest
```

## 使用方法

```bash
# 扫描当前目录
depx scan

# 扫描指定目录
depx scan /path/to/project
```

输出示例：

```
  Project Summary
--------------------------
  Path:            /path/to/project
  Package Manager: npm
  Dependencies:    12
  Used:            7
  Unused:          5

  Unused Dependencies
--------------------------
  [x] moment
  [x] chalk
  [x] typescript
```

## 架构

```
depx
├── cmd/                    # CLI 入口
│   ├── root.go            # 根命令
│   └── scan.go            # scan 子命令
├── internal/
│   ├── analyzer/          # 分析器 - 协调扫描流程
│   │   └── unused_test.go # 集成测试
│   ├── manifest/          # 清单解析 - 读取依赖声明
│   │   ├── npm.go         # package.json 解析
│   │   ├── gomod.go       # go.mod 解析
│   │   └── manifest_test.go # 清单解析测试
│   ├── usage/             # 使用分析 - 检测代码引用
│   │   ├── js.go          # JS/TS import 分析
│   │   ├── golang.go      # Go import 分析
│   │   ├── js_test.go     # JS 分析器单元测试
│   │   ├── golang_test.go # Go 分析器单元测试
│   │   └── boundary_test.go # 边界条件测试
│   └── report/            # 报告生成 - 格式化输出
├── testdata/              # 测试夹具数据
│   ├── npm-project/       # npm 基础项目
│   ├── npm-complex/       # npm 复杂项目（Vue/TS/hooks）
│   ├── go-project/        # Go 基础项目
│   ├── go-complex/        # Go 复杂项目（多包）
│   ├── edge-all-used/     # 边界：全部使用
│   ├── edge-none-used/    # 边界：全部未使用
│   ├── edge-no-source/    # 边界：无源码
│   └── real-npm/          # npm 真实场景模拟
└── main.go
```

## 技术实现

- **语言**: Go
- **CLI 框架**: cobra
- **输出着色**: fatih/color
- **依赖检测**: 正则匹配 + 状态机注释过滤

核心流程：

1. 检测项目类型（npm/go）
2. 解析清单文件获取依赖列表
3. 遍历源码文件提取 import 语句
4. 过滤注释和字符串字面量
5. 匹配依赖声明与实际使用
6. 生成报告

## 限制

- 仅检测直接依赖，不分析传递依赖
- npm 项目中 `@types/*` 包始终显示为未使用（由 TypeScript 编译器自动加载）
- Go 项目中自动排除 `// indirect` 标记的间接依赖

## License

MIT
