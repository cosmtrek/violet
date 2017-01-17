package index

import (
	"github.com/cosmtrek/violet/pkg/analyzer"
)

var (
	gsegmenter *analyzer.Segmenter
)

func segmenter() *analyzer.Segmenter {
	if gsegmenter == nil {
		segmenter, _ := analyzer.New()
		gsegmenter = segmenter
	}
	return gsegmenter
}

// MergeDocIDs merges docs in non-decreasing order
func MergeDocIDs(a []Doc, b []Doc) ([]Doc, bool) {
	aLen := len(a)
	bLen := len(b)
	if aLen == 0 && bLen == 0 {
		return nil, false
	}

	var i, j, k int
	c := make([]Doc, len(a)+len(b))
	for i < aLen && j < bLen {
		if a[i] == b[j] {
			c[k] = a[i]
			i++
			j++
			k++
			continue
		}
		if a[i].DocID < b[j].DocID {
			c[k] = a[i]
			i++
			k++
		} else {
			c[k] = b[j]
			j++
			k++
		}
	}

	if i < aLen {
		for i < aLen {
			c[k] = a[i]
			i++
			k++
		}
	} else {
		for j < bLen {
			c[k] = b[j]
			j++
			k++
		}
	}

	return c[:k], true
}

// IntersectDocIDs returns the intersections of two doc ids
func IntersectDocIDs(a []Doc, b []Doc) ([]Doc, bool) {
	aLen := len(a)
	bLen := len(b)
	if aLen == 0 && bLen == 0 {
		return nil, false
	}

	var i, j, k int
	var cLen int
	if aLen > bLen {
		cLen = aLen
	} else {
		cLen = bLen
	}
	c := make([]Doc, cLen)
	for i < aLen && j < bLen {
		if a[i] == b[j] {
			c[k] = a[i]
			i++
			j++
			k++
			continue
		}
		if a[i].DocID < b[j].DocID {
			i++
		} else {
			j++
		}
	}
	return c[:k], true
}
