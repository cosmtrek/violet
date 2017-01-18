package main

import (
	"flag"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/cosmtrek/violet/engine/api"
)

var (
	indexPath string
	index     string
	fields    string
	dataFile  string
	query     string
)

func main() {
	log.Println("violet is pretty!")

	flag.StringVar(&indexPath, "path", "", "path")
	flag.StringVar(&index, "index", "violet", "a string")
	flag.StringVar(&fields, "fields", "", "field1-type,field2-type,field3-type")
	flag.StringVar(&dataFile, "data", "", "path")
	flag.StringVar(&query, "query", "", "term;date>10|len>2")
	flag.Parse()

	if indexPath == "" {
		log.Errorln("must set index path")
		os.Exit(1)
	}
	if fields == "" {
		log.Errorln("must set field meta")
		os.Exit(1)
	}
	if dataFile == "" {
		log.Errorln("must set data file")
		os.Exit(1)
	}

	var indexer *api.Indexer
	var err error
	fieldsMeta := make(map[string]uint64, 0)
	var fieldsArr []string
	for _, f := range strings.Split(fields, ",") {
		fs := strings.Split(f, "-")
		fs1, _ := strconv.Atoi(fs[1])
		fieldsMeta[fs[0]] = uint64(fs1)
		fieldsArr = append(fieldsArr, fs[0])
	}
	indexer, err = api.NewIndexer(indexPath, nil)
	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
	if err = indexer.AddIndex(index, fieldsMeta); err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
	if err = indexer.LoadDocumentsFromFile(index, dataFile, "text", fieldsArr); err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
	log.Println("load data from file successfully!")
	docs, ok := indexer.Search(index, query, nil)
	if ok {
		log.Println(docs)
	} else {
		log.Println("no result")
	}
}
