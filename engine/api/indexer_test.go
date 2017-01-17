package api

import (
	"testing"

	"github.com/cosmtrek/violet/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewIndexer(t *testing.T) {
	path, err := utils.TempDir("", true)
	assert.Nil(t, err)
	indexer, err := NewIndexer(path, nil)
	assert.Nil(t, err)
	assert.NotNil(t, indexer)
}
