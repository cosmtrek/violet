package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyze(t *testing.T) {
	segmenter, err := New()
	assert.NotNil(t, segmenter)
	assert.Nil(t, err)

	terms := segmenter.Analyze("violet is a search engine in go!", false)
	expected1 := []string{"violet", "is", "a", "search", "engine", "in", "go"}
	assert.Equal(t, expected1, terms)

	termsWithSearch := segmenter.Analyze("violet is a search engine in go!", true)
	expected2 := []string{"violet", "is", "a", "search", "engine", "in", "go"}
	assert.Equal(t, expected2, termsWithSearch)
}
