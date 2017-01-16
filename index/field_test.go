package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewField(t *testing.T) {
	path, err := tempDir("", true)
	assert.Nil(t, err)
	field, err := NewField("fieldA", TString, path, segmenter())
	assert.Nil(t, err)
	assert.NotNil(t, field)
}

func TestField_addDocument_searchTerm_getDetail(t *testing.T) {
	path, err := tempDir("", true)
	assert.Nil(t, err)
	field, err := NewField("fieldB", TString, path, segmenter())
	assert.Nil(t, err)
	field.addDocument(uint64(0), "若你遇到他 蔡健雅 红色高跟鞋")
	field.addDocument(uint64(1), "该怎么去形容你最贴切")
	field.addDocument(uint64(2), "拿什么跟你作比较才算特别")
	field.addDocument(uint64(3), "对你的感觉 强烈")
	field.addDocument(uint64(4), "却又不太了解 只凭直觉")
	field.addDocument(uint64(5), "你像我在被子里的舒服")
	field.addDocument(uint64(6), "却又像风捉摸不住")
	field.addDocument(uint64(7), "像手腕上散发的香水味")
	field.addDocument(uint64(8), "像爱不释手的 红色高跟鞋")
	field.addDocument(uint64(9), "该怎么去形容你最贴切")
	field.addDocument(uint64(10), "拿什么跟你作比较才算特别")
	field.addDocument(uint64(11), "对你的感觉 强烈")
	field.addDocument(uint64(12), "却又不太了解 只凭直觉")
	field.addDocument(uint64(13), "你像我在被子里的舒服")
	field.addDocument(uint64(14), "却又像风捉摸不住")
	field.addDocument(uint64(15), "像手腕上散发的香水味")
	field.addDocument(uint64(16), "像爱不释手的 红色高跟鞋")
	field.addDocument(uint64(17), "你像我在被子里的舒服")
	field.addDocument(uint64(18), "却又像风捉摸不住")
	field.addDocument(uint64(19), "像手腕上散发的香水味")
	field.addDocument(uint64(20), "像爱不释手的 红色高跟鞋")
	field.addDocument(uint64(21), "我爱你有种左灯右行的冲突")
	field.addDocument(uint64(22), "疯狂却怕没有退路")
	field.addDocument(uint64(23), "你能否让我停止这种追逐")
	field.addDocument(uint64(24), "就这么双 最后唯一的")
	field.addDocument(uint64(25), "红色高跟鞋")
	err = field.syncToDisk()
	assert.Nil(t, err)
	docs, found := field.searchTerm("香水")
	assert.True(t, found)
	expected := []Doc{{DocID: 7}, {DocID: 15}, {DocID: 19}}
	assert.EqualValues(t, expected, docs)

	doc7 := "像手腕上散发的香水味"
	expectedDoc7, _, foundDoc, err := field.getDetail(uint64(7))
	assert.True(t, foundDoc)
	assert.Nil(t, err)
	assert.Equal(t, expectedDoc7, doc7)
}
