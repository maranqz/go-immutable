package immutable

// Tests for linters.

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/tools/go/analysis/analysistest"
)

type linterSuite struct {
	suite.Suite
}

func (suite *linterSuite) TestContextLinter() {
	analysistest.Run(suite.T(), TestdataDir(),
		ImmutableAnalyzer, "testlintdata/scalar")
}

func TestLinterSuite(t *testing.T) {
	suite.Run(t, new(linterSuite))
}

func TestdataDir() string {
	_, testFilename, _, ok := runtime.Caller(1)
	if !ok {
		panic("unable to get current test filename")
	}

	return filepath.Join(filepath.Dir(testFilename), "testdata")
}
