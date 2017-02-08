package index

import (
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

const (
	// NULL default value of filter type field
	NULL = iota
	// EQUAL =
	EQUAL
	// LESS <
	LESS
	// GREATER >
	GREATER
	// EXCLUDE not contains this term
	EXCLUDE
)

// Filter on fields
type Filter struct {
	Field string
	Value interface{}
	Ftype uint64
}

// TermFilter small unit for query string
type TermFilter struct {
	Term   string
	Filter Filter
}

// Query for searching
type Query struct {
	Index   *Index
	Content string
}

// NewQuery initializes a query
func NewQuery(index *Index, query string) (*Query, error) {
	if index == nil || query == "" {
		return nil, errors.New("invalid params")
	}
	q := &Query{
		Index:   index,
		Content: strings.TrimSpace(query),
	}
	return q, nil
}

func (q *Query) do() ([]Doc, bool) {
	termFilters, err := q.analyzeQuery()
	if err != nil {
		log.Error(err)
		return nil, false
	}
	var wanted []TermFilter
	var excluded []TermFilter
	var numfiltered []TermFilter
	for _, tf := range termFilters {
		if tf.Filter.Ftype == EXCLUDE {
			excluded = append(excluded, tf)
			continue
		}
		if tf.Filter.Ftype == LESS || tf.Filter.Ftype == EQUAL || tf.Filter.Ftype == GREATER {
			numfiltered = append(numfiltered, tf)
		} else {
			wanted = append(wanted, tf)
		}
	}

	var docs []Doc
	for i := range wanted {
		subdocs, found := q.search(wanted[i])
		if found {
			docs = append(docs, subdocs...)
		}
	}
	for i := range excluded {
		subdocs, found := q.search(excluded[i])
		if found {
			rest, found := ExcludeDocIDs(docs, subdocs)
			if found {
				docs = rest
			}
			// TODO ignore error?
		}
	}
	if len(numfiltered) > 0 {
		var ndocs []Doc
		for _, doc := range docs {
			filtered := false
			for _, f := range numfiltered {
				value, ok := f.Filter.Value.(uint64)
				if ok {
					if !q.Index.Fields[f.Filter.Field].filter(doc.DocID, value, f.Filter.Ftype) {
						filtered = true
						break
					}
				}
			}
			if !filtered {
				ndocs = append(ndocs, doc)
			}
		}
		if len(ndocs) == 0 {
			return nil, false
		}
		return ndocs, true
	}
	return docs, true
}

func (q *Query) search(termFilter TermFilter) ([]Doc, bool) {
	var docs []Doc
	filter := termFilter.Filter
	terms := q.Index.Segmenter.Analyze(termFilter.Term, false)
	if filter.Field == "*" {
		// no constraints
		first := true
		for _, term := range terms {
			var subdocs []Doc
			for k, v := range q.Index.FieldMeta {
				if v == TString {
					fieldDocs, ok := q.Index.SearchTerm(term, k)
					if ok {
						subdocs, _ = MergeDocIDs(subdocs, fieldDocs)
					}
				}
			}
			if first {
				docs = subdocs
				first = false
			} else {
				docs, _ = IntersectDocIDs(docs, subdocs)
			}
		}
		return docs, true
	}
	// search single field
	first := true
	// invalid field
	ftype, ok := q.Index.FieldMeta[filter.Field]
	if !ok {
		return nil, false
	}
	for _, term := range terms {
		var subdocs []Doc
		if ftype == TString {
			fieldDocs, ok := q.Index.SearchTerm(term, filter.Field)
			if ok {
				subdocs, _ = MergeDocIDs(subdocs, fieldDocs)
			}
		}
		if first {
			docs = subdocs
			first = false
		} else {
			docs, _ = IntersectDocIDs(docs, subdocs)
		}
	}
	return docs, true
}

func (q *Query) analyzeQuery() ([]TermFilter, error) {
	if len(q.Content) == 0 {
		return nil, nil
	}
	segs := strings.Split(q.Content, " ")
	var tfs []TermFilter
	for _, seg := range segs {
		tf := new(TermFilter)
		// search "len>5"
		operator, isCompare := HasCompare(seg)
		if isCompare {
			segkv := strings.Split(seg, operator)
			tf.Term = segkv[0]
			value, err := strconv.Atoi(segkv[1])
			if err != nil {
				// TODO handle error?
				continue
			}
			switch operator {
			case "<":
				tf.Filter = Filter{Field: tf.Term, Value: uint64(value), Ftype: LESS}
			case "=":
				tf.Filter = Filter{Field: tf.Term, Value: uint64(value), Ftype: EQUAL}
			case ">":
				tf.Filter = Filter{Field: tf.Term, Value: uint64(value), Ftype: GREATER}
			default:
			}
			if tf != nil {
				tfs = append(tfs, *tf)
			}
			continue
		}
		segkv := strings.Split(seg, ":")
		if len(segkv) == 1 {
			// search "-word"
			if segkv[0][0] == '-' {
				tf.Term = segkv[0][1:]
				tf.Filter = Filter{Field: "*", Value: tf.Term, Ftype: EXCLUDE}
			} else {
				// search "word"
				tf.Term = segkv[0]
				tf.Filter = Filter{Field: "*", Value: tf.Term}
			}
		} else if len(segkv) == 2 {
			// search "-field:word"
			if segkv[0][0] == '-' {
				tf.Term = segkv[0][1:]
				tf.Filter = Filter{Field: tf.Term, Value: segkv[1], Ftype: EXCLUDE}
			} else {
				// search "field:word"
				tf.Term = segkv[0]
				tf.Filter = Filter{Field: tf.Term, Value: segkv[1]}
			}
		} else {
			return nil, errors.New("failed to analyze query")
		}
		if tf != nil {
			tfs = append(tfs, *tf)
		}
	}
	return tfs, nil
}
