package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_tempDir(t *testing.T) {
	uniqueFile, err := tempDir("", true)
	assert.Nil(t, err)
	assert.NotNil(t, uniqueFile)

	file, err := tempDir("", false)
	assert.Nil(t, err)
	assert.NotNil(t, file)
}
