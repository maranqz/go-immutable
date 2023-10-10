package immutable

// Tests for linters.

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/packages"
)

// TODO remove.
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// TODO remove.
func TestTryDST(t *testing.T) {
	t.Parallel()

	pp, _ := decorator.Load(&packages.Config{
		//nolint:staticcheck
		Mode:       packages.LoadAllSyntax,
		Context:    nil,
		Logf:       nil,
		Dir:        "local_path",
		Env:        nil,
		BuildFlags: nil,
		Fset:       nil,
		ParseFile:  nil,
		Tests:      false,
		Overlay:    nil,
	}, "./...")

	_ = pp
}

func TestLinterSuite(t *testing.T) {
	testdata := analysistest.TestData()

	tests := map[string]struct {
		pkgs []string
	}{
		"scalar":               {pkgs: []string{"scalar"}},
		"global":               {pkgs: []string{"global"}},
		"structs_local":        {pkgs: []string{"structs/local"}},
		"struct_only_exported": {pkgs: []string{"structs/only_exported/..."}},
		"struct_global":        {pkgs: []string{"structs/global/..."}},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			dirs := make([]string, 0, len(tt.pkgs))

			for _, pkg := range tt.pkgs {
				dirs = append(dirs, filepath.Join(testdata, "src", pkg))
			}

			analysistest.Run(t, TestdataDir(),
				ImmutableAnalyzer, dirs...)
		})
	}
}

func TestdataDir() string {
	_, testFilename, _, ok := runtime.Caller(1)
	if !ok {
		panic("unable to get current test filename")
	}

	return filepath.Join(filepath.Dir(testFilename), "testdata")
}
