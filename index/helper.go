package index

import (
	"io/ioutil"
	"os"

	libgo "github.com/cosmtrek/libgo/utils"
	"github.com/cosmtrek/violet/pkg/analyzer"
)

var (
	gsegmenter *analyzer.Segmenter
)

func segmenter() *analyzer.Segmenter {
	if gsegmenter == nil {
		segmenter, _ := analyzer.New()
		gsegmenter = segmenter
	}
	return gsegmenter
}

func tempDir(prefix string, random bool) (string, error) {
	dir, err := ioutil.TempDir(os.TempDir(), prefix)
	if err != nil {
		return "", err
	}
	if random {
		dir, err = ioutil.TempDir(os.TempDir(), prefix+libgo.RandomString(20))
		if err != nil {
			return "", err
		}
	}
	return dir, nil
}
