package testscript

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "scripts",
		Setup: func(env *testscript.Env) error {
			root, err := moduleRoot()
			if err != nil {
				return err
			}

			binDir := filepath.Join(env.WorkDir, "bin")
			if err := os.MkdirAll(binDir, 0o755); err != nil {
				return fmt.Errorf("create bin dir: %w", err)
			}

			exePath := filepath.Join(binDir, "github-janitor")
			build := exec.Command("go", "build", "-o", exePath, ".")
			build.Dir = root
			out, err := build.CombinedOutput()
			if err != nil {
				return fmt.Errorf("build github-janitor: %w\n%s", err, out)
			}

			path := env.Getenv("PATH")
			env.Setenv("PATH", binDir+string(os.PathListSeparator)+path)
			return nil
		},
	})
}

func moduleRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime.Caller failed")
	}
	// This file lives at test/testscript/testscript_test.go.
	// The module root is two levels above test/.
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../..")), nil
}
