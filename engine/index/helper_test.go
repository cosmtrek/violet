package index

import (
	"testing"

	"github.com/cosmtrek/violet/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func Test_tempDir(t *testing.T) {
	uniqueFile, err := utils.TempDir("", true)
	assert.Nil(t, err)
	assert.NotNil(t, uniqueFile)

	file, err := utils.TempDir("", false)
	assert.Nil(t, err)
	assert.NotNil(t, file)
}

func TestMergeDocIDs(t *testing.T) {
	a := []Doc{
		{DocID: 0},
		{DocID: 1},
		{DocID: 2},
		{DocID: 3},
		{DocID: 4},
		{DocID: 5},
	}
	b := []Doc{
		{DocID: 2},
		{DocID: 4},
		{DocID: 5},
		{DocID: 6},
		{DocID: 8},
	}
	expected := []Doc{
		{DocID: 0},
		{DocID: 1},
		{DocID: 2},
		{DocID: 3},
		{DocID: 4},
		{DocID: 5},
		{DocID: 6},
		{DocID: 8},
	}
	actual, found := MergeDocIDs(a, b)
	assert.True(t, found)
	assert.EqualValues(t, expected, actual)
}

func TestIntersectDocIDs(t *testing.T) {
	a := []Doc{
		{DocID: 0},
		{DocID: 1},
		{DocID: 2},
		{DocID: 3},
		{DocID: 4},
	}
	b := []Doc{
		{DocID: 2},
		{DocID: 3},
		{DocID: 4},
	}
	expected := []Doc{
		{DocID: 2},
		{DocID: 3},
		{DocID: 4},
	}
	actual, found := IntersectDocIDs(a, b)
	assert.True(t, found)
	assert.EqualValues(t, expected, actual)
}

func TestExcludeDocIDs(t *testing.T) {
	a := []Doc{
		{DocID: 2},
		{DocID: 3},
		{DocID: 4},
		{DocID: 6},
		{DocID: 7},
	}
	b := []Doc{
		{DocID: 0},
		{DocID: 2},
	}
	expected := []Doc{
		{DocID: 3},
		{DocID: 4},
		{DocID: 6},
		{DocID: 7},
	}
	actual, found := ExcludeDocIDs(a, b)
	assert.True(t, found)
	assert.EqualValues(t, expected, actual)
}
