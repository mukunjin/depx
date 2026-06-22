package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mukunjin/depx/internal/analyzer"
	"github.com/mukunjin/depx/internal/lockfile"
	"github.com/mukunjin/depx/internal/manifest"
	"github.com/mukunjin/depx/internal/report"
	"github.com/mukunjin/depx/internal/surface"
	"github.com/spf13/cobra"
)

var (
	surfaceConfigPath string
	surfaceIndirect   bool
	surfaceDev        bool
)

// typesPrefix 是 TypeScript 类型包的统一前缀
const typesPrefix = "@types/"

var surfaceCmd = &cobra.Command{
	Use:   "surface [path]",
	Short: "Analyze dependency surface area",
	Long: `Analyze how widely each dependency is used across the project.

By default only runtime dependencies are analyzed. Use --dev to include devDependencies.
Use --indirect to show a summary of shared transitive dependencies from the lock file.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		return runSurface(path, surfaceConfigPath, surfaceIndirect, surfaceDev)
	},
}

func runSurface(path, configPath string, showIndirect, includeDev bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path %s: %w", path, err)
	}

	cfg, err := loadConfig(absPath, configPath)
	if err != nil {
		return err
	}

	m, err := analyzer.DetectManifest(absPath)
	if err != nil {
		return err
	}

	deps, err := m.Dependencies()
	if err != nil {
		return fmt.Errorf("reading dependencies: %w", err)
	}
	if includeDev {
		deps, err = manifest.MergeWithDev(m)
		if err != nil {
			return fmt.Errorf("reading dependencies: %w", err)
		}
	}

	var filteredDeps []string
	for _, dep := range deps {
		if cfg.IsIgnored(dep) {
			continue
		}
		if strings.HasPrefix(dep, typesPrefix) {
			continue
		}
		filteredDeps = append(filteredDeps, dep)
	}

	directSet := make(map[string]struct{}, len(filteredDeps))
	for _, dep := range filteredDeps {
		directSet[dep] = struct{}{}
	}

	opts := &surface.Options{
		ManifestType:    m.Type(),
		ExcludeDirs:     cfg.ExcludeDirs,
		ExcludeFiles:    cfg.ExcludeFiles,
		ReadNodeModules: cfg.ReadNodeModules,
	}

	runtimeResults, err := surface.AnalyzeSurface(absPath, filteredDeps, opts)
	if err != nil {
		return err
	}

	var indirectTotal int
	var parentCounts map[string]int
	if showIndirect {
		if !cfg.LockFile {
			fmt.Fprintln(os.Stderr, "Warning: lock_file is disabled in config; --indirect has no effect")
		} else {
			lf, err := lockfile.DetectLockFile(absPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not detect lockfile: %v\n", err)
			} else {
				lockDeps, err := lf.Dependencies()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not read lockfile dependencies: %v\n", err)
				} else {
					indirectSet := make(map[string]struct{})
					for _, dep := range lockDeps {
						if _, isDirect := directSet[dep.Name]; isDirect {
							continue
						}
						if strings.HasPrefix(dep.Name, typesPrefix) {
							continue
						}
						indirectSet[dep.Name] = struct{}{}
					}
					indirectTotal = len(indirectSet)

					if lfNpm, ok := lf.(*lockfile.NpmLockFile); ok {
						parentCounts = lfNpm.SharedIndirectCounts(filteredDeps)
					}
				}
			}
		}
	} // 输出报告
	report.PrintSurfaceReport(runtimeResults, indirectTotal, parentCounts)
	return nil
}

func init() {
	surfaceCmd.Flags().StringVarP(&surfaceConfigPath, "config", "c", "", "Path to config file (.depx.yml)")
	surfaceCmd.Flags().BoolVarP(&surfaceIndirect, "indirect", "i", false, "Show shared indirect dependency summary")
	surfaceCmd.Flags().BoolVarP(&surfaceDev, "dev", "D", false, "Include devDependencies (default: runtime dependencies only)")
	rootCmd.AddCommand(surfaceCmd)
}
