package router

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/luscis/openvrr/pkg/api"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Oops!", http.StatusNotFound)
}

func NotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Oops!", http.StatusMethodNotAllowed)
}

type Http struct {
	listen     string
	adminToken string
	adminFile  string
	server     *http.Server
	url        *mux.Router
	caller     api.Caller
}

func (h *Http) Init() {
	h.SetToken()
	r := h.Url()
	if h.server == nil {
		h.server = &http.Server{
			Addr:         h.listen,
			Handler:      r,
			ReadTimeout:  5 * time.Minute,
			WriteTimeout: 10 * time.Minute,
		}
	}
	h.AddUrl()
}

func (h *Http) IsAuth(w http.ResponseWriter, r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false
	}

	if user == "vrr" && pass == h.adminToken {
		return true
	}

	return false
}

func (h *Http) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Http.Middleware %s %s", r.Method, r.URL.Path)
		if h.IsAuth(w, r) {
			next.ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic")
			http.Error(w, "Authorization Required", http.StatusUnauthorized)
		}
	})
}

func (h *Http) Url() *mux.Router {
	if h.url == nil {
		h.url = mux.NewRouter()
		h.url.NotFoundHandler = http.HandlerFunc(NotFound)
		h.url.MethodNotAllowedHandler = http.HandlerFunc(NotAllowed)
		h.url.Use(h.Middleware)
	}

	return h.url
}

func (h *Http) AddUrl() {
	url := h.Url()

	url.HandleFunc("/api/urls", h.GetApi).Methods("GET")
	api.Add(url, h.caller)
}

func (h *Http) saveToken(token string) {
	f, err := os.OpenFile(h.adminFile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("Http.saveToken: %s", err)
		return
	}
	defer f.Close()
	if _, err := f.Write([]byte(token)); err != nil {
		log.Printf("Http.saveToken: %s", err)
		return
	}
}

func (h *Http) SetToken() {
	token := ""
	if _, err := os.Stat(h.adminFile); os.IsNotExist(err) {
		log.Printf("Http.SetToken: file:%s does not exist", h.adminFile)
	} else {
		contents, err := os.ReadFile(h.adminFile)
		if err != nil {
			log.Printf("Http.SetToken: file:%s %s", h.adminFile, err)
		} else {
			token = strings.TrimSpace(string(contents))
		}
	}
	if token == "" {
		token = api.GenString(32)
		h.saveToken(token)
	}
	h.adminToken = token
}

func (t *Http) GetApi(w http.ResponseWriter, r *http.Request) {
	var urls []string
	t.url.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil || !strings.HasPrefix(path, "/api") {
			return nil
		}
		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}
		for _, m := range methods {
			urls = append(urls, fmt.Sprintf("%-6s %s", m, path))
		}
		return nil
	})
	api.ResponseYaml(w, urls)
}

func (h *Http) Start() {
	log.Printf("Http.Start %s", h.listen)

	go func() {
		if err := h.server.ListenAndServe(); err != nil {
			log.Printf("Http.Start on %s: %s", h.listen, err)
			return
		}
	}()
}
