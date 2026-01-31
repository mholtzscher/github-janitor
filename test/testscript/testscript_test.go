package testscript

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "scripts",
	})
}
