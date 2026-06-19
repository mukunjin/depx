package cmd

import (
	"testing"
)

func TestRootCommandBasic(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
	if rootCmd.Use != "depx" {
		t.Errorf("expected Use 'depx', got %q", rootCmd.Use)
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	expectedCmds := map[string]bool{
		"scan":    false,
		"surface": false,
	}

	for _, cmd := range rootCmd.Commands() {
		if _, ok := expectedCmds[cmd.Name()]; ok {
			expectedCmds[cmd.Name()] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestVersionVariable(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	// 默认值是 "dev"，通过 build.ps1 构建时会被 Git tag 覆盖（如 "v0.2.0"）
}

func TestScanCommandFlags(t *testing.T) {
	flag := scanCmd.Flags().Lookup("config")
	if flag == nil {
		t.Error("scan command should have --config flag")
	}
	if flag.Shorthand != "c" {
		t.Errorf("expected shorthand 'c', got %q", flag.Shorthand)
	}
}

func TestSurfaceCommandFlags(t *testing.T) {
	flag := surfaceCmd.Flags().Lookup("config")
	if flag == nil {
		t.Error("surface command should have --config flag")
	}
	if flag.Shorthand != "c" {
		t.Errorf("expected shorthand 'c', got %q", flag.Shorthand)
	}
}

func TestRootCommandUseAndShort(t *testing.T) {
	if rootCmd.Use != "depx" {
		t.Errorf("expected Use 'depx', got %q", rootCmd.Use)
	}
	// 验证 Short 描述不为空
	if rootCmd.Short == "" {
		t.Error("root command Short description should not be empty")
	}
	// 验证 Long 描述不为空
	if rootCmd.Long == "" {
		t.Error("root command Long description should not be empty")
	}
}

func TestExecute(t *testing.T) {
	// 测试 Execute 函数不返回错误（在没有参数时）
	// 为避免测试框架传入的测试标志污染命令参数，使用 SetArgs 隔离参数
	rootCmd.SetArgs([]string{})
	err := Execute()
	if err != nil {
		t.Errorf("Execute() should not return error, got: %v", err)
	}
}
