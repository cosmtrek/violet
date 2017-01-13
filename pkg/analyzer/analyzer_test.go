package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyze(t *testing.T) {
	analyzer, err := New()
	assert.NotNil(t, analyzer)
	assert.Nil(t, err)

	segments := analyzer.Analyze("violet is a search engine in go!")
	expected := []string{"violet", "is", "a", "search", "engine", "in", "go"}
	assert.Equal(t, expected, segments)
}
