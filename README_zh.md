# depx

[![Go 版本](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![许可证](https://img.shields.io/badge/License-GPLv3-blue?style=flat-square)](./LICENSE)

依赖分析工具 - 检测项目中未使用的依赖，分析依赖影响面。

**English Documentation**: [README.md](README.md)

---

## 功能

- **未使用检测** — 扫描项目依赖声明文件，检测未使用的依赖
- **影响面分析** — 分析依赖在项目中的使用广度，评估关键度
- **Lock File 分析** — 解析 Lock File 获取准确的依赖版本并检测间接依赖
- **配置支持** — 通过 `.depx.yml` 自定义忽略规则、排除目录

## 支持范围

| 包管理器 | 清单文件 | Lock File | 源文件 |
|---------|----------|-----------|--------|
| npm | package.json | package-lock.json | .js, .ts, .jsx, .tsx, .mjs, .cjs, .vue, .svelte |
| Go | go.mod | go.sum | .go |
| Rust | Cargo.toml | Cargo.lock | .rs |
| Python | requirements.txt | — | .py |

## 安装

### 首次设置（仅 Windows）

Windows 默认禁止运行 PowerShell 脚本。请以**管理员身份**运行 PowerShell 并执行：

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

输入 `Y` 确认。此设置永久生效。

### 安装 depx

```powershell
# 克隆仓库
git clone https://github.com/mukunjin/depx.git
cd depx

# 构建（自动从 Git tag 获取版本）
.\build.ps1

# 安装
.\install.ps1
```

### 卸载

```powershell
.\install.ps1 -Uninstall
```

脚本会自动完成：
- 将 `depx.exe` 复制到 `%LOCALAPPDATA%\depx`
- 添加到用户 PATH
- 提示重启终端使配置生效

### 版本管理

版本号由以下位置控制：

| 位置 | 作用 | 生效时机 |
|------|------|----------|
| Git tag | 主要来源 | 通过 `.\build.ps1` 构建时 |
| `cmd/root.go` 第 8 行 | 回退值（`dev`） | 直接运行 `go build` 且不使用 `-ldflags` 时 |
| `build.ps1` | 读取 git tag，通过 `-ldflags` 注入 | 每次运行 `.\build.ps1` 时 |
| `install.ps1` 第 139 行 | 从二进制读取版本（`depx --version`） | 安装验证时 |

**工作原理：**
1. `build.ps1` 运行 `git describe --tags --abbrev=0` 获取最新的 Git tag
2. tag 通过 `-ldflags="-X github.com/mukunjin/depx/cmd.Version=<tag>"` 注入到二进制
3. `cmd/root.go` 提供回退值（`dev`），当不使用 `-ldflags` 时生效
4. `install.ps1` 通过读取版本验证安装的二进制

**如何发布新版本：**
1. 创建 Git tag：`git tag v0.3.0`
2. 推送 tag：`git push origin v0.3.0`
3. 运行 `.\build.ps1` 构建新版本
4. 运行 `.\install.ps1` 安装

**查看当前版本：**
```powershell
depx --version
```

## 使用方法

### 扫描

```bash
# 扫描当前目录
depx scan

# 扫描指定目录
depx scan C:\path\to\project

# 使用自定义配置扫描
depx scan --config C:\path\to\.depx.yml

# 显示帮助
depx --help

# 显示版本
depx --version
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

### 影响面分析

```bash
# 分析依赖影响面
depx surface
```

输出示例：

```
  Dependency Surface Area
--------------------------
  axios
    Files: 43
    Modules: 8
    References: 182
    Criticality: High

  chalk
    Files: 4
    Modules: 1
    References: 6
    Criticality: Low
```

## 配置

在项目根目录创建 `.depx.yml`：

```yaml
# 忽略特定依赖
ignore:
  - "@types/node"
  - "typescript"

# 排除目录
exclude_dirs:
  - "vendor"
  - "dist"
  - "node_modules"

# 排除文件模式
exclude_files:
  - "*.test.js"
  - "*.spec.ts"

# 读取 node_modules 进行精确分析
read_node_modules: false

# 启用 Lock File 分析
lock_file: true
```

## 技术实现

- **语言**: Go
- **CLI 框架**: cobra
- **输出着色**: fatih/color
- **YAML 解析**: gopkg.in/yaml.v3
- **依赖检测**: 正则匹配 + 状态机注释过滤

核心流程：

1. 检测项目类型（npm/go/cargo/pip）
2. 解析清单文件获取依赖列表
3. 解析 Lock File（如果可用）
4. 从 `.depx.yml` 加载配置
5. 遍历源码文件提取 import 语句
6. 过滤注释和字符串字面量
7. 匹配依赖声明与实际使用
8. 分析影响面
9. 生成报告

## 限制

- 仅检测直接依赖，不分析传递依赖
- npm 项目中 `@types/*` 包始终显示为未使用（TypeScript 编译器自动加载）
- Go 项目中自动排除 `// indirect` 标记的间接依赖
- Python 包名可能与 import 名不一致（如 `pip install Pillow` → `import PIL`）

## 许可证

GPLv3 - 详见 [LICENSE](LICENSE)。
