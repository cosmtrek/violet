package analyzer

import (
	"os"

	"github.com/huichen/sego"
	"github.com/pkg/errors"
)

const (
	stopwords = "!@#$%^&*()_+=-`~,./<>?;':\"[]{}，。！￥……（）·；：「」『』、？《》】【“”|\\的 "
)

// Analyzer is used to segment words
type Analyzer struct {
	segmenter sego.Segmenter
	stopword  map[string]bool
}

// New initializes analyzer
func New() (*Analyzer, error) {
	var analyzer Analyzer
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	dict := wd + "/data/dictionary.txt"
	if _, err := os.Stat(dict); os.IsNotExist(err) {
		return nil, errors.New("Dictionary is not found")
	}
	analyzer.segmenter.LoadDictionary(dict)
	analyzer.stopword = make(map[string]bool, 80)
	swrune := []rune(stopwords)
	for _, w := range swrune {
		analyzer.stopword[string(w)] = true
	}
	return &analyzer, nil
}

// Analyze return valid words that not contains stop words
func (a *Analyzer) Analyze(text string) []string {
	if text == "" {
		return []string{}
	}
	segments := sego.SegmentsToSlice(a.segmenter.Segment([]byte(text)), false)
	var validSegs []string
	for _, seg := range segments {
		if _, ok := a.stopword[seg]; !ok {
			validSegs = append(validSegs, seg)
		}
	}
	return validSegs
}
