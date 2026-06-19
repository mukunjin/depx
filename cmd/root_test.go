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
