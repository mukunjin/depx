package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand(t *testing.T) {
	// 测试根命令执行
	cmd := rootCmd
	if cmd == nil {
		t.Fatal("rootCmd is nil")
	}

	if cmd.Use != "depx" {
		t.Errorf("expected Use to be 'depx', got '%s'", cmd.Use)
	}

	if cmd.Short != "Dependency efficiency analyzer" {
		t.Errorf("expected Short to be 'Dependency efficiency analyzer', got '%s'", cmd.Short)
	}

	if cmd.Version != "dev" {
		t.Errorf("expected Version to be 'dev', got '%s'", cmd.Version)
	}
}

func TestRootCommandHelp(t *testing.T) {
	// 测试帮助信息
	cmd := rootCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected help output, got empty string")
	}
}

func TestRootCommandVersion(t *testing.T) {
	// 测试版本信息
	cmd := rootCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected version output, got empty string")
	}
}

func TestExecute(t *testing.T) {
	// 测试 Execute 函数
	err := Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
