package api

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"time"

	"gopkg.in/yaml.v2"
)

func ResponseJson(w http.ResponseWriter, v interface{}) {
	str, err := json.Marshal(v)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ResponseYaml(w http.ResponseWriter, v interface{}) {
	str, err := yaml.Marshal(v)
	if err == nil {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetData(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return err
	}
	return nil
}

func GetQueryOne(req *http.Request, name string) string {
	query := req.URL.Query()
	if values, ok := query[name]; ok {
		return values[0]
	}
	return ""
}

var Letters = []byte("0123456789abcdefghijklmnopqrstuvwxyz")

func GenString(n int) string {
	buffer := make([]byte, n)

	rr := rand.New(rand.NewSource(time.Now().Unix()))
	for i := range buffer {
		buffer[i] = Letters[rr.Int63()%int64(len(Letters))]
	}
	buffer[0] = Letters[rr.Int63()%26+10]

	return string(buffer)
}
