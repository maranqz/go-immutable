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

	pp, err := decorator.Load(&packages.Config{
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
	check(err)

	_ = pp
}

func TestLinterSuite(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()

	tests := []struct {
		pkg string
	}{
		{pkg: "scalar"},
		{pkg: "global/..."},
		{pkg: "struct/..."},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.pkg, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join(testdata, "src", tt.pkg)

			analysistest.Run(t, TestdataDir(),
				ImmutableAnalyzer, dir)
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
