package index

import (
	"fmt"
	"strconv"

	"github.com/cosmtrek/violet/pkg/io"
	"github.com/cosmtrek/violet/pkg/utils"
	"github.com/pkg/errors"
)

// Source holds documents in order
type Source struct {
	maxDocID  uint64
	filepath  string
	field     string
	fieldType uint64
	handler   *io.Mmap
	detail    *io.Mmap
}

// NewSource initialize source struct
func NewSource(filepath string, field string, fieldType uint64) (*Source, error) {
	source := &Source{
		maxDocID:  0,
		filepath:  filepath,
		field:     field,
		fieldType: fieldType,
	}

	sourceFilename := fmt.Sprintf("%v%v.source", filepath, field)
	detailFilename := fmt.Sprintf("%v%v.detail", filepath, field)

	var err error
	if fieldType == TString || fieldType == TStore {
		if utils.FileExists(sourceFilename) && utils.FileExists(detailFilename) {
			if source.handler, err = io.NewMmap(sourceFilename, io.ModeAppend); err != nil {
				return nil, errors.Wrap(err, "failed to handle source file for string and store field in append mode")
			}
			if source.detail, err = io.NewMmap(detailFilename, io.ModeAppend); err != nil {
				return nil, errors.Wrap(err, "failed to handle detail file for string and store field in append mode")
			}
		} else {
			if source.handler, err = io.NewMmap(sourceFilename, io.ModeCreate); err != nil {
				return nil, errors.Wrap(err, "failed to create source file for string and store field ")
			}
			if source.detail, err = io.NewMmap(detailFilename, io.ModeCreate); err != nil {
				return nil, errors.Wrap(err, "failed to create detail file for string and store field ")
			}
		}
	}

	if fieldType == TNumber {
		if utils.FileExists(sourceFilename) {
			if source.handler, err = io.NewMmap(sourceFilename, io.ModeAppend); err != nil {
				return nil, errors.Wrap(err, "failed to handle source file for number field in append mode")
			}
		} else {
			if source.handler, err = io.NewMmap(sourceFilename, io.ModeCreate); err != nil {
				return nil, errors.Wrap(err, "failed to create source file for number field")
			}
		}
		source.detail = nil
	}
	return source, nil
}

func (s *Source) addDocument(docid uint64, content string) error {
	var err error
	if s.fieldType == TString || s.fieldType == TStore {
		offset := uint64(s.detail.GetPointer())
		if err = s.handler.AppendUint64(offset); err != nil {
			return errors.Wrap(err, "failed to append offset to source file")
		}
		s.handler.Sync()
		if err = s.detail.AppendStringWithLen(content); err != nil {
			return errors.Wrap(err, "failed to append string to detail file")
		}
		s.detail.Sync()
		return nil
	}

	if s.fieldType == TNumber {
		val, err := strconv.ParseUint(content, 10, 64)
		if err != nil {
			val = 0
		}
		if err = s.handler.AppendUint64(val); err != nil {
			return errors.Wrap(err, "failed to append uint64 number to source file")
		}
		return nil
	}
	return nil
}

func (s *Source) getDetail(docid uint64) interface{} {
	offset := s.handler.ReadUint64(docid * 8)
	if s.fieldType == TString || s.fieldType == TStore {
		return s.detail.ReadStringWithLen(offset)
	}
	return offset
}

func (s *Source) sync() error {
	var err error
	if s.fieldType == TString || s.fieldType == TStore {
		if err = s.detail.Sync(); err != nil {
			return err
		}
	}
	return s.handler.Sync()
}

func (s *Source) string() string {
	return fmt.Sprintf("[SOURCE] maxdocid: %d, filepath: %s, field: %s, fieldType: %d, handler: %s, detail: %s",
		s.maxDocID, s.filepath, s.field, s.fieldType, s.handler.String(), s.detail.String())
}
