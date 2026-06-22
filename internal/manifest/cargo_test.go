package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCargoManifest(t *testing.T) {
	tmpDir := t.TempDir()

	cargoToml := `[package]
name = "my-app"
version = "0.1.0"

[dependencies]
serde = "1.0"
tokio = { version = "1.0", features = ["full"] }
reqwest = "0.11"

[dev-dependencies]
mockall = "0.11"
`

	cargoPath := filepath.Join(tmpDir, "Cargo.toml")
	if err := os.WriteFile(cargoPath, []byte(cargoToml), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := NewCargoManifest(tmpDir)
	if err != nil {
		t.Fatalf("NewCargoManifest failed: %v", err)
	}

	if manifest.Type() != "cargo" {
		t.Errorf("Expected type 'cargo', got '%s'", manifest.Type())
	}

	deps, err := manifest.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 3 {
		t.Errorf("Expected 3 runtime dependencies, got %d: %v", len(deps), deps)
	}

	devDeps, err := manifest.DevDependencies()
	if err != nil {
		t.Fatalf("DevDependencies failed: %v", err)
	}
	if len(devDeps) != 1 || devDeps[0] != "mockall" {
		t.Errorf("Expected dev dependency mockall, got %v", devDeps)
	}

	expected := map[string]bool{"serde": true, "tokio": true, "reqwest": true}
	for _, dep := range deps {
		if !expected[dep] {
			t.Errorf("Unexpected dependency: %s", dep)
		}
	}
}

func TestCargoManifestMissing(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := NewCargoManifest(tmpDir)
	if err == nil {
		t.Error("Expected error for missing Cargo.toml")
	}
}

func TestCargoManifestEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	cargoToml := `[package]
name = "empty"
version = "0.1.0"
`

	cargoPath := filepath.Join(tmpDir, "Cargo.toml")
	if err := os.WriteFile(cargoPath, []byte(cargoToml), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := NewCargoManifest(tmpDir)
	if err != nil {
		t.Fatalf("NewCargoManifest failed: %v", err)
	}

	deps, err := manifest.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(deps))
	}
}

func TestCargoManifestComments(t *testing.T) {
	tmpDir := t.TempDir()

	cargoToml := `[package]
name = "test"

[dependencies]
# This is a comment
serde = "1.0"
# Another comment
tokio = "1.0"
`

	cargoPath := filepath.Join(tmpDir, "Cargo.toml")
	if err := os.WriteFile(cargoPath, []byte(cargoToml), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := NewCargoManifest(tmpDir)
	if err != nil {
		t.Fatalf("NewCargoManifest failed: %v", err)
	}

	deps, err := manifest.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d: %v", len(deps), deps)
	}
}
