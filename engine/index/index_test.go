package index

import (
	"strconv"
	"testing"

	"github.com/cosmtrek/violet/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewIndex(t *testing.T) {
	index, err := NewIndex("index", "violet", segmenter())
	assert.Nil(t, err)
	assert.NotNil(t, index)
}

func TestIndex_IndexFields_AddDocument_Search_GetDocument(t *testing.T) {
	path, err := utils.TempDir("", true)
	assert.Nil(t, err)
	index, err := NewIndex(path, "violet", segmenter())
	assert.Nil(t, err)
	assert.NotNil(t, index)
	fieldMap := map[string]uint64{
		"a": TString,
		"b": TNumber,
	}
	err = index.IndexFields(fieldMap)
	assert.Nil(t, err)
	song := []string{
		"其实很简单 其实很自然",
		"两个人的爱由两人分担",
		"其实并不难 是你太悲观",
		"隔着一道墙不跟谁分享",
		"不想让你为难",
		"你不再需要给我个答案",
		"我想你是爱我的",
		"我猜你也舍不得",
		"但是怎么说 总觉得",
		"我们之间留了太多空白格",
		"也许你不是我的",
		"爱你却又该割舍",
		"分开或许是选择",
		"但它也可能是我们的缘分",
		"其实很简单 其实很自然",
		"两个人的爱由两人分担",
		"其实并不难 是你太悲观",
		"隔着一道墙不跟说分享",
		"不想让你为难",
		"你不再需要给我个答案",
		"我想你是爱我的",
		"我猜你也舍不得",
		"但是怎么说 总觉得",
		"我们之间留了太多空白格",
		"也许你不是我的",
		"爱你却又该割舍",
		"分开或许是选择",
		"但它也可能是我们的缘分",
		"我想你是爱我的",
		"我猜你也舍不得",
		"但是怎么说 总觉得",
		"我们之间留了太多空白格",
		"也许你不是我的",
		"爱你却又该割舍",
		"分开或许是选择",
		"但它也可能是我们的缘分",
	}
	for i := range song {
		doc := make(map[string]string, 2)
		doc["a"] = song[i]
		doc["b"] = strconv.Itoa(len(song[i]))
		err = index.AddDocument(doc)
		assert.Nil(t, err)
	}
	err = index.SyncToDisk()
	assert.Nil(t, err)
	filters1 := []Filter{
		{Field: "b", Value: uint64(10), Ftype: GREATER},
	}
	docs1, found1 := index.Search("我们之间留了太多空白格", filters1)
	assert.True(t, found1)
	expected1 := []Doc{{DocID: 9}, {DocID: 23}, {DocID: 31}}
	assert.EqualValues(t, expected1, docs1)
	filters2 := []Filter{
		{Field: "b", Value: uint64(50), Ftype: GREATER},
	}
	_, found2 := index.Search("我们之间留了太多空白格", filters2)
	assert.False(t, found2)
}
