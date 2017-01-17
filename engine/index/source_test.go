package index

import (
	"testing"

	"github.com/cosmtrek/violet/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewSource(t *testing.T) {
	var err error
	path, err := utils.TempDir("", true)
	assert.Nil(t, err)
	_, err = NewSource(path, "field1", TString)
	assert.Nil(t, err)
	_, err = NewSource(path, "field2", TStore)
	assert.Nil(t, err)
	_, err = NewSource(path, "field3", TNumber)
	assert.Nil(t, err)
}

func TestSource_addDocument(t *testing.T) {
	path, err := utils.TempDir("", true)
	assert.Nil(t, err)
	s, err := NewSource(path, "field1", TString)
	assert.Nil(t, err)
	err = s.addDocument(uint64(0), "doc content")
	assert.Nil(t, err)
}

func TestSource_getDetail(t *testing.T) {
	path, err := utils.TempDir("", true)
	assert.Nil(t, err)
	s1, err := NewSource(path, "field1", TString)
	assert.Nil(t, err)
	docid1, doc1 := uint64(0), "doc content 1"
	err = s1.addDocument(docid1, doc1)
	assert.Nil(t, err)
	assert.Equal(t, doc1, s1.getDetail(docid1))

	s2, err := NewSource(path, "field2", TStore)
	assert.Nil(t, err)
	docid2, doc2 := uint64(0), "doc content 22"
	err = s2.addDocument(docid2, doc2)
	assert.Nil(t, err)
	assert.Equal(t, doc2, s2.getDetail(docid2))

	s3, err := NewSource(path, "field3", TNumber)
	assert.Nil(t, err)
	docid3, doc3 := uint64(0), "1"
	err = s3.addDocument(docid3, doc3)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), s3.getDetail(docid3))
}
