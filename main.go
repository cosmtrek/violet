package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/cosmtrek/violet/engine/api"
	"github.com/pressly/chi"
)

var (
	indexPath  string
	index      string
	fields     string
	dataFile   string
	query      bool
	serverMode bool
	serverPort string

	handler *api.Handler
)

func main() {
	log.Println("violet is pretty!")

	flag.StringVar(&indexPath, "path", "", "path")
	flag.StringVar(&index, "index", "violet", "a string")
	flag.StringVar(&fields, "fields", "", "field1-type,field2-type,field3-type")
	flag.StringVar(&dataFile, "data", "", "path")
	flag.BoolVar(&query, "query", false, "term;date>10|len>2")
	flag.BoolVar(&serverMode, "server", true, "server mode")
	flag.StringVar(&serverPort, "port", "6000", "server port")
	flag.Parse()

	var err error
	handler = new(api.Handler)
	if serverMode {
		r := chi.NewRouter()
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})
		r.Post("/index", handler.IndexHandler)
		r.Get("/:indexer/search", handler.SearchHandler)
		log.Fatal(http.ListenAndServe(":"+serverPort, r))
	}

	indexer := handler.Indexer
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

	if query {
		input := bufio.NewScanner(os.Stdin)
		fmt.Println("> Enter query:")
		var term string
		for input.Scan() {
			term = input.Text()
			if term == "q" || term == "quit" {
				break
			}
			// TODO add filters
			docs, ok := indexer.Search(index, term, nil)
			fmt.Println("~ results ")
			if ok {
				for i, d := range docs {
					fmt.Printf("%d, %s at %s\n", i, d["tweet"], d["date"])
				}
			} else {
				fmt.Println("no result")
			}
			fmt.Println("> Enter query:")
		}
		log.Println("goodbye, my friend~")
	}
}
