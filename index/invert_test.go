package index

import (
	"os"
	"testing"

	"github.com/cosmtrek/violet/pkg/analyzer"
	"github.com/stretchr/testify/assert"
)

func TestNewInvert(t *testing.T) {
	dir := os.TempDir()
	analyzer, err := analyzer.New()
	assert.Nil(t, err)
	invert, err := NewInvert(dir, "field", TString, analyzer)
	assert.Nil(t, err)
	assert.NotNil(t, invert)
}
