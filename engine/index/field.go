package index

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/cosmtrek/violet/pkg/analyzer"
	"github.com/cosmtrek/violet/pkg/utils"
	"github.com/kurrik/json"
	"github.com/pkg/errors"
)

// Field holds source file and invert file
type Field struct {
	Name     string `json:"name"`
	Type     uint64 `json:"type"`
	MaxDocID uint64 `json:"maxdocid"`
	Path     string
	source   *Source
	invert   *Invert
}

// NewField initializes a field struct
func NewField(name string, ftype uint64, path string, segmenter analyzer.Analyzer) (*Field, error) {
	field := &Field{
		Name: name,
		Type: ftype,
		Path: path,
	}
	var err error
	metafile := fieldMetaFile(path, name)
	if utils.FileExists(metafile) {
		meta, err := utils.ReadJSON(metafile)
		if err != nil {
			log.Errorf("failed to read json file %s, err: %s\n", metafile, err.Error())
			return nil, errors.Wrap(err, "failed to read json file")
		}
		if err = json.Unmarshal(meta, field); err != nil {
			log.Errorf("failed to unmarshal meta %v into field %v, err: %s\n", meta, field, err.Error())
			return nil, errors.Wrap(err, "failed to unmarshal meta into field")
		}
	}
	if field.source, err = NewSource(path, name, ftype); err != nil {
		log.Errorf("failed to create source file, err: %s\n", err.Error())
		return nil, errors.Wrap(err, "failed to create source file")
	}
	if ftype == TString || ftype == TStore {
		if field.invert, err = NewInvert(path, name, ftype, segmenter); err != nil {
			log.Errorf("failed to create invert file, err: %s\n", err.Error())
			return nil, errors.Wrap(err, "failed to create invert file")
		}
	}
	return field, nil
}

// addDocument adds document into source file and invert file that synced into disk at intervals
func (f *Field) addDocument(docid uint64, doc string) error {
	if docid != f.MaxDocID {
		return errors.New("docid is not equal current max docid")
	}
	var err error
	if err = f.source.addDocument(docid, doc); err != nil {
		return errors.Wrap(err, "failed to add document into source file")
	}
	if f.invert != nil {
		if err = f.invert.addDocument(docid, doc); err != nil {
			return errors.Wrap(err, "failed to add document into invert file")
		}
	}
	f.MaxDocID++
	if f.MaxDocID%InvertSyncInterval == 0 {
		if f.invert != nil {
			if err = f.invert.saveTmpInvert(); err != nil {
				return errors.Wrap(err, "failed to save tmp invert file to disk")
			}
		}
	}
	return nil
}

// getDetail returns string if field type is TString or TStore, uint64 if field type is TNumber
func (f *Field) getDetail(docid uint64) (string, uint64, bool, error) {
	if docid > f.MaxDocID || f.source == nil {
		return "", 0, false, nil
	}
	val := f.source.getDetail(docid)
	if f.Type == TString || f.Type == TStore {
		return fmt.Sprintf("%s", val), 0, true, nil
	}
	num, ok := val.(uint64)
	if !ok {
		return "", 0, false, errors.New("failed to type asserting for uint64")
	}
	return "", num, true, nil
}

func (f *Field) searchTerm(term string) ([]Doc, bool) {
	if f.invert != nil {
		return f.invert.searchTerm(term)
	}
	return nil, false
}

func (f *Field) filter(docid, value, ftype uint64) bool {
	// current only support number type
	if f.source == nil || f.Type != TNumber {
		return false
	}
	return f.source.filter(docid, value, ftype)
}

func (f *Field) syncToDisk() error {
	var err error
	if f.invert != nil {
		if err = f.invert.saveTmpInvert(); err != nil {
			return errors.Wrap(err, "failed to sync tmp invert file to disk")
		}
		if err = f.invert.mergeTmpInvert(); err != nil {
			return errors.Wrap(err, "failed to merge tmp invert files")
		}
	}
	if f.source != nil {
		if err = f.source.sync(); err != nil {
			return errors.Wrap(err, "failed to sync source file to disk")
		}
	}
	file := fieldMetaFile(f.Path, f.Name)
	if err = utils.WriteJSON(file, f); err != nil {
		return errors.Wrap(err, "failed to write field into json")
	}
	return nil
}

func fieldMetaFile(filepath, field string) string {
	return fmt.Sprintf("%v%v.json", filepath, field)
}

func (f *Field) String() string {
	return fmt.Sprintf("[FIELD] name: %s, type: %s, maxdocid: %d, path: %s, source: %s, invert: %s",
		f.Name, f.Type, f.MaxDocID, f.Path, f.source.string(), f.invert.string())
}
