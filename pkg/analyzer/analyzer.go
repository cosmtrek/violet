package analyzer

import (
	"os"

	"github.com/huichen/sego"
	"github.com/pkg/errors"
)

const (
	stopwords = "!@#$%^&*()_+=-`~,./<>?;':\"[]{}，。！￥……（）·；：「」『』、？《》】【“”|\\的 "
)

// Analyzer exposed
type Analyzer interface {
	Analyze(text string, searchMode bool) []string
}

// Segmenter is used to segment words
type Segmenter struct {
	handler  sego.Segmenter
	stopword map[string]bool
}

// New initializes analyzer
func New() (*Segmenter, error) {
	var segmenter Segmenter
	dict := os.Getenv("violet") + "/pkg/analyzer/data/dictionary.txt"
	print(dict)
	if _, err := os.Stat(dict); os.IsNotExist(err) {
		return nil, errors.New("Dictionary is not found")
	}
	segmenter.handler.LoadDictionary(dict)
	segmenter.stopword = make(map[string]bool, 80)
	swrune := []rune(stopwords)
	for _, w := range swrune {
		segmenter.stopword[string(w)] = true
	}
	return &segmenter, nil
}

// Analyze return valid words that not contains stop words
func (s *Segmenter) Analyze(text string, searchModel bool) []string {
	if text == "" {
		return []string{}
	}
	segments := sego.SegmentsToSlice(s.handler.Segment([]byte(text)), searchModel)
	var validSegs []string
	for _, seg := range segments {
		if _, ok := s.stopword[seg]; !ok {
			validSegs = append(validSegs, seg)
		}
	}
	return validSegs
}
