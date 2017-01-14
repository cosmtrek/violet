package index

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSource(t *testing.T) {
	tmpDir := os.TempDir()
	var err error

	_, err = NewSource(tmpDir, "field1", TString)
	assert.Nil(t, err)
	_, err = NewSource(tmpDir, "field2", TStore)
	assert.Nil(t, err)
	_, err = NewSource(tmpDir, "field3", TNumber)
	assert.Nil(t, err)
}

func Test_addDocument(t *testing.T) {
	tmpDir := os.TempDir()

	s, err := NewSource(tmpDir, "field1", TString)
	assert.Nil(t, err)
	err = s.addDocument(uint64(0), "doc content")
	assert.Nil(t, err)
}

func Test_getDetail(t *testing.T) {
	tmpDir := os.TempDir()

	s1, err := NewSource(tmpDir, "field1", TString)
	assert.Nil(t, err)
	docid1, doc1 := uint64(0), "doc content 1"
	err = s1.addDocument(docid1, doc1)
	assert.Nil(t, err)
	assert.Equal(t, doc1, s1.getDetail(docid1))

	s2, err := NewSource(tmpDir, "field2", TStore)
	assert.Nil(t, err)
	docid2, doc2 := uint64(0), "doc content 22"
	err = s2.addDocument(docid2, doc2)
	assert.Nil(t, err)
	assert.Equal(t, doc2, s2.getDetail(docid2))

	s3, err := NewSource(tmpDir, "field3", TNumber)
	assert.Nil(t, err)
	docid3, doc3 := uint64(0), "1"
	err = s3.addDocument(docid3, doc3)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), s3.getDetail(docid3))
}
