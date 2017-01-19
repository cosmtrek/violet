package index

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmtrek/violet/pkg/analyzer"
	"github.com/cosmtrek/violet/pkg/utils"
	"github.com/pkg/errors"
)

const (
	// TString string type
	TString = iota
	// TNumber number type
	TNumber
	// TStore store type
	TStore
)

// Index is the entry to all low level data structures
type Index struct {
	Name      string            `json:"index"`
	MaxDocID  uint64            `json:"maxdocid"`
	Path      string            `json:"path"`
	FieldMeta map[string]uint64 `json:"fields"`
	Fields    map[string]*Field
	Segmenter analyzer.Analyzer
}

// NewIndex initializes index
func NewIndex(path string, name string, segmenter analyzer.Analyzer) (*Index, error) {
	index := &Index{
		Name:      name,
		Path:      path,
		FieldMeta: nil,
		Segmenter: segmenter,
		Fields:    make(map[string]*Field),
	}
	metafile := indexMetaFile(path, name)
	if utils.FileExists(metafile) {
		meta, err := utils.ReadJSON(metafile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read json file")
		}
		if err = json.Unmarshal(meta, index); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal meta into index")
		}
		fieldPath := fmt.Sprintf("%v/%v_", path, name)
		for fname, ftype := range index.FieldMeta {
			field, err := NewField(fname, ftype, fieldPath, segmenter)
			if err != nil {
				return nil, errors.Wrap(err, "failed to load field file")
			}
			index.Fields[fname] = field
		}
	} else {
		// TODO
	}
	return index, nil
}

// IndexFields documents field's meta info
func (x *Index) IndexFields(fields map[string]uint64) error {
	if x.FieldMeta == nil {
		x.FieldMeta = fields
		fieldPath := fmt.Sprintf("%v/%v_", x.Path, x.Name)
		for fname, ftype := range fields {
			field, err := NewField(fname, ftype, fieldPath, x.Segmenter)
			if err != nil {
				return errors.Wrap(err, "failed to create field file")
			}
			x.Fields[fname] = field
		}
		return nil
	}
	return errors.New("fields existed")
}

// AddDocument inserts documents to index
func (x *Index) AddDocument(doc map[string]string) error {
	if x.FieldMeta == nil {
		return errors.New("no field meta")
	}
	docid := x.MaxDocID
	x.MaxDocID++
	for name, field := range x.Fields {
		if _, ok := doc[name]; !ok {
			doc[name] = ""
		}
		if err := field.addDocument(docid, doc[name]); err != nil {
			x.MaxDocID--
			return err
		}
	}
	return nil
}

// Search query and returns docs
func (x *Index) Search(query string, filters []Filter) ([]Doc, bool) {
	terms := x.Segmenter.Analyze(query, false)
	var docs []Doc
	first := true
	for _, term := range terms {
		var subdocs []Doc
		for k, v := range x.FieldMeta {
			if v == TString {
				fieldDocs, ok := x.SearchTerm(term, k)
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

	// filter doc ids
	if len(filters) > 0 {
		var fdocs []Doc
		for _, doc := range docs {
			filtered := false
			for _, f := range filters {
				if !x.Fields[f.Field].filter(doc.DocID, f.Value, f.Ftype) {
					filtered = true
					break
				}
			}
			if !filtered {
				fdocs = append(fdocs, doc)
			}
		}
		if len(fdocs) == 0 {
			return nil, false
		}
		return fdocs, true
	}

	if len(docs) == 0 {
		return nil, false
	}
	return docs, true
}

// SearchTerm returns docs that contains term
func (x *Index) SearchTerm(term, field string) ([]Doc, bool) {
	t := strings.TrimSpace(term)
	if len(t) <= 0 {
		return nil, false
	}
	return x.Fields[field].searchTerm(term)
}

// GetDocument returns document source
func (x *Index) GetDocument(docid uint64) (map[string]string, bool) {
	if docid > x.MaxDocID {
		return nil, false
	}
	doc := make(map[string]string)
	for fname, field := range x.Fields {
		v, _, ok, err := field.getDetail(docid)
		if err != nil {
			return nil, false
		}
		if ok {
			doc[fname] = v
		} else {
			doc[fname] = ""
		}
	}
	doc["docid"] = strconv.Itoa(int(docid))
	return doc, true
}

// SyncToDisk flushes documents into disk
func (x *Index) SyncToDisk() error {
	if x.FieldMeta == nil {
		return errors.New("no field meta")
	}
	var err error
	for _, field := range x.Fields {
		if err = field.syncToDisk(); err != nil {
			return err
		}
	}
	return nil
}

func indexMetaFile(filepath, index string) string {
	return fmt.Sprintf("%v/%v.json", filepath, index)
}
