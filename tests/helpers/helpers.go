package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mukunjin/depx/internal/analyzer"
)

// ScanTestdata scans the testdata directory and returns the analyzer result.
func ScanTestdata(t *testing.T, dir string) *analyzer.ScanResult {
	t.Helper()
	testdataDir := filepath.Join("..", "testdata", dir)
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", testdataDir)
	}

	result, err := analyzer.Scan(testdataDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	return result
}

// AssertScanResult verifies basic integrity of a ScanResult.
func AssertScanResult(t *testing.T, result *analyzer.ScanResult, expectedType string) {
	t.Helper()
	if result.ManifestType != expectedType {
		t.Errorf("Expected manifest type '%s', got '%s'", expectedType, result.ManifestType)
	}
	if result.TotalDeps == 0 {
		t.Error("Expected some dependencies, got 0")
	}
	if result.Path == "" {
		t.Error("Expected non-empty path")
	}
	if result.UsageDetails == nil {
		t.Error("Expected non-nil UsageDetails")
	}
	if result.UsedDeps+result.UnusedDeps != result.TotalDeps {
		t.Errorf("Stats mismatch: Used(%d) + Unused(%d) != Total(%d)", result.UsedDeps, result.UnusedDeps, result.TotalDeps)
	}
}
