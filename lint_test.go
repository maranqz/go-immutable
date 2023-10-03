package immutable

// Tests for linters.

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestLinterSuite(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()

	tests := []struct {
		pkg string
	}{
		{pkg: "scalar"},
		{pkg: "global"},
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
