package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/pressly/chi"
)

// Handler for apis
type Handler struct {
	Indexer *Indexer
}

// IndexerRequest creates an indexer
type IndexerRequest struct {
	Index     string `json:"index"`
	IndexPath string `json:"index_path"`
	Datafile  string `json:"datafile"`
	Fields    string `json:"fields"`
}

// Response returns message to client
type Response struct {
	Code    string              `json:"code"`
	Status  string              `json:"status"`
	Message string              `json:"message,omitempty"`
	Docs    []map[string]string `json:"docs,omitempty"`
}

// IndexHandler creates an indexer
func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	var request IndexerRequest
	var err error
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorln(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(responseFailed("1", "failed to read request body"))
		return
	}
	if err = json.Unmarshal(body, &request); err != nil {
		log.Errorln(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(responseFailed("1", "failed to unmarshal request body"))
		return
	}

	fieldsMeta := make(map[string]uint64, 0)
	var fieldsArr []string
	for _, f := range strings.Split(request.Fields, ",") {
		fs := strings.Split(f, "-")
		fs1, _ := strconv.Atoi(fs[1])
		fieldsMeta[fs[0]] = uint64(fs1)
		fieldsArr = append(fieldsArr, fs[0])
	}
	h.Indexer, err = NewIndexer(request.IndexPath, nil)
	if err != nil {
		log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(responseFailed("1", err.Error()))
		return
	}
	if err = h.Indexer.AddIndex(request.Index, fieldsMeta); err != nil {
		log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(responseFailed("1", err.Error()))
		return
	}
	if err = h.Indexer.LoadDocumentsFromFile(request.Index, request.Datafile, "text", fieldsArr); err != nil {
		log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(responseFailed("1", err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(responseOk("created indexer successfully"))
}

// SearchHandler searches everything via http
func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	if h.Indexer == nil {
		w.WriteHeader(http.StatusOK)
		w.Write(responseFailed("2", "please create indexer firstly"))
		return
	}
	indexer := chi.URLParam(r, "indexer")
	query := r.URL.Query().Get("query")

	docs, ok := h.Indexer.Search(indexer, query, nil)
	if ok {
		resp := Response{
			Code:   "0",
			Status: "OK",
			Docs:   docs,
		}
		data, err := json.Marshal(resp)
		if err != nil {
			log.Errorln(err)
			w.Write([]byte("{}"))
		} else {
			w.Write(data)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(responseOk("no data"))
}

func responseFailed(code string, msg string) []byte {
	resp := Response{
		Code:    code,
		Status:  "FAILED",
		Message: msg,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Errorln(err)
		return []byte("{}")
	}
	return data
}

func responseOk(msg string) []byte {
	resp := Response{
		Code:    "0",
		Status:  "OK",
		Message: msg,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Errorln(err)
		return []byte("{}")
	}
	return data
}
