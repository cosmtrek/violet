package index

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/cosmtrek/violet/pkg/analyzer"
	"github.com/cosmtrek/violet/pkg/io"
	"github.com/cosmtrek/violet/pkg/skeleton"
	"github.com/cosmtrek/violet/pkg/utils"
	"github.com/pkg/errors"
)

const (
	// InvertSyncInterval save invert files at intervals
	InvertSyncInterval = 100000
)

// Invert is the core of search engine
type Invert struct {
	curDocID   uint64
	field      string
	fieldType  uint64
	filepath   string
	tmpIvts    []tmpIvt
	idx        *io.Mmap
	terms      skeleton.KV
	segmentNum uint64
	segmenter  analyzer.Analyzer
	wg         *sync.WaitGroup
}

type tmpMergeTable struct {
	Term string
	Docs []Doc
}

// NewInvert initializes invert struct
func NewInvert(filepath string, field string, fieldtype uint64, segmenter analyzer.Analyzer) (*Invert, error) {
	ivt := &Invert{
		filepath:  filepath,
		field:     field,
		fieldType: fieldtype,
		segmenter: segmenter,
		wg:        new(sync.WaitGroup),
	}
	var err error
	idxfile := idxFile(filepath, field)
	dicfile := dicFile(filepath, field)
	ivt.terms = skeleton.NewHashMap()
	if utils.FileExists(idxfile) && utils.FileExists(dicfile) {
		if err = ivt.terms.Load(idxfile); err != nil {
			return nil, errors.Wrap(err, "failed to load idx file")
		}
		idx, err := io.NewMmap(idxfile, io.ModeAppend)
		if err != nil {
			return nil, errors.Wrap(err, "failed to mmap idx file")
		}
		ivt.idx = idx
	} else {
		idx, err := io.NewMmap(idxfile, io.ModeCreate)
		if err != nil {
			return nil, errors.Wrap(err, "failed to mmap idx file")
		}
		ivt.idx = idx
	}

	return ivt, nil
}

func (v *Invert) addDocument(docid uint64, content string) error {
	terms := v.segmenter.Analyze(content, true)
	for _, term := range terms {
		t := strings.TrimSpace(term)
		if len(t) > 0 {
			// prevent duplicated tmpIvt
			found := false
			for _, e := range v.tmpIvts {
				if e.Term == t && e.DocID == docid {
					found = true
					break
				}
			}
			if !found {
				v.tmpIvts = append(v.tmpIvts, tmpIvt{DocID: docid, Term: t})
			}
		}
	}
	return nil
}

func (v *Invert) saveTmpInvert() error {
	file := fmt.Sprintf("%v%v_%v.ivt", v.filepath, v.field, v.segmentNum)
	v.segmentNum++
	sort.Sort(TmpIvtTermSort(v.tmpIvts))

	fd, err := os.Create(file)
	if err != nil {
		return errors.Wrap(err, "failed to create "+file)
	}
	defer fd.Close()

	for _, tv := range v.tmpIvts {
		tvj, err := json.Marshal(tv)
		if err != nil {
			return errors.Wrap(err, "failed to marshal tmpivt into json")
		}
		fd.WriteString(string(tvj) + "\n")
	}
	v.tmpIvts = make([]tmpIvt, 0)
	return nil
}

func (v *Invert) mergeTmpInvert() error {
	var tableChans []chan tmpMergeTable
	for i := uint64(0); i < v.segmentNum; i++ {
		file := ivtFile(v.filepath, v.field, i)
		tableChans = append(tableChans, make(chan tmpMergeTable))
		v.wg.Add(1)
		go v.mapRoutine(file, &tableChans[i])
	}
	v.wg.Add(1)
	file := ivtFile(v.filepath, v.field, v.segmentNum)
	go v.reduceRoutine(file, &tableChans)
	v.wg.Wait()
	return nil
}

func (v *Invert) mapRoutine(file string, tableChan *chan tmpMergeTable) error {
	defer v.wg.Done()

	fd, err := os.Open(file)
	if err != nil {
		return errors.Wrap(err, "failed to open file in mapRoutine")
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	var table tmpMergeTable
	if scanner.Scan() {
		var ivt tmpIvt
		content := scanner.Text()
		json.Unmarshal([]byte(content), &ivt)
		table.Term = ivt.Term
		table.Docs = make([]Doc, 0)
		table.Docs = append(table.Docs, Doc{DocID: ivt.DocID})
	}
	for scanner.Scan() {
		var ivt tmpIvt
		content := scanner.Text()
		json.Unmarshal([]byte(content), &ivt)
		if ivt.Term == table.Term {
			table.Docs = append(table.Docs, Doc{DocID: ivt.DocID})
		} else {
			*tableChan <- table
			table.Term = ivt.Term
			table.Docs = make([]Doc, 0)
			table.Docs = append(table.Docs, Doc{DocID: ivt.DocID})
		}
	}
	*tableChan <- table
	close(*tableChan)
	os.Remove(file)
	return nil
}

func (v *Invert) reduceRoutine(file string, tableChans *[]chan tmpMergeTable) error {
	defer v.wg.Done()

	idxfile := idxFile(v.filepath, v.field)
	idxFd, err := os.OpenFile(idxfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to open idx file in reduceRoutine")
	}
	defer idxFd.Close()

	dicfile := dicFile(v.filepath, v.field)
	dicFd, err := os.OpenFile(dicfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to open dic file in reduceRoutine")
	}
	defer dicFd.Close()

	tableLens := len(*tableChans)
	closeFlag := make([]bool, tableLens)
	var tables []tmpMergeTable
	v.terms = skeleton.NewHashMap()
	var maxTerm string
	var offsetTotal uint64

	for i, t := range *tableChans {
		tt, ok := <-t
		if ok {
			if maxTerm < tt.Term {
				maxTerm = tt.Term
			}
		} else {
			closeFlag[i] = true
		}
		tables = append(tables, tt)
	}

	var nextMax string
	for {
		var restable tmpMergeTable
		restable.Docs = make([]Doc, 0)
		restable.Term = maxTerm
		closeNum := 0
		for i := range tables {
			if maxTerm == tables[i].Term {
				restable.Docs = append(restable.Docs, tables[i].Docs...)
				tt, ok := <-(*tableChans)[i]
				if ok {
					tables[i].Term = tt.Term
					tables[i].Docs = tt.Docs
				} else {
					closeFlag[i] = true
				}
			}
			if !closeFlag[i] {
				if nextMax <= tables[i].Term {
					nextMax = tables[i].Term
				}
				closeNum++
			}
		}
		sort.Sort(DocSort(restable.Docs))
		docsLen := uint64(len(restable.Docs))
		lenBuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(lenBuf, docsLen)
		idxFd.Write(lenBuf)

		buf := new(bytes.Buffer)
		if err = binary.Write(buf, binary.LittleEndian, restable.Docs); err != nil {
			return err
		}
		idxFd.Write(buf.Bytes())
		v.terms.Push(restable.Term, uint64(offsetTotal))
		offsetTotal = offsetTotal + uint64(8) + docsLen*8
		if closeNum == 0 {
			break
		}
		maxTerm = nextMax
		nextMax = ""
	}
	v.terms.Save(dicfile)
	return v.reloadIvtFileAfterMerge(dicfile)
}

func (v *Invert) reloadIvtFileAfterMerge(filename string) error {
	return v.terms.Load(filename)
}

func (v *Invert) searchTerm(term string) ([]Doc, bool) {
	t := strings.TrimSpace(term)
	if len(t) == 0 {
		return nil, false
	}
	if offset, ok := v.terms.Get(t); ok {
		valLen := v.idx.ReadInt64(int64(offset))
		docs := readDocIDs(v.idx, uint64(offset+8), uint64(valLen))
		return docs, true
	}
	return nil, false
}

type tmpIvt struct {
	Term  string `json:"term"`
	DocID uint64 `json:"docid"`
}

// TmpIvtTermSort sorts tmpIvt array
type TmpIvtTermSort []tmpIvt

func (t TmpIvtTermSort) Len() int           { return len(t) }
func (t TmpIvtTermSort) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TmpIvtTermSort) Less(i, j int) bool { return t[i].Term > t[j].Term }

// Doc means docid
type Doc struct {
	DocID uint64 `json:"docid`
}

// DocSort sorts doc array
type DocSort []Doc

func (d DocSort) Len() int           { return len(d) }
func (d DocSort) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d DocSort) Less(i, j int) bool { return d[i].DocID < d[j].DocID }

func idxFile(filepath, field string) string {
	return fmt.Sprintf("%v%v.idx", filepath, field)
}

func dicFile(filepath, field string) string {
	return fmt.Sprintf("%v%v.dic", filepath, field)
}

func ivtFile(filepath string, field string, num uint64) string {
	return fmt.Sprintf("%v%v_%v.ivt", filepath, field, num)
}

func readDocIDs(m *io.Mmap, start uint64, idsLen uint64) []Doc {
	return *(*[]Doc)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&m.MmapBytes[start])),
		Len:  int(idsLen),
		Cap:  int(idsLen),
	}))
}

func (v *Invert) string() string {
	return fmt.Sprintf("[INVERT] field: %s, fieldType: %d, filepath: %s, idx: %s, segmentNum: %d",
		v.field, v.fieldType, v.filepath, v.idx.String(), v.segmentNum)
}
