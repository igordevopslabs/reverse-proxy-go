package main

import (
	"bytes"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type ReverseProxy struct {
	routes map[string][]string
}

func NewReverseProxy() *ReverseProxy {
	return &ReverseProxy{
		routes: map[string][]string{
			"/todos/1": {
				"https://jsonplaceholder.typicode.com",
				"https://jsonplaceholder.typicode.com"},
		},
	}
}

func (rp *ReverseProxy) selectBackend(path string) (string, bool) {
	backend, exists := rp.routes[path]

	if !exists || len(backend) == 0 {
		log.Printf("there is no backend or path defined")
		return "", false
	}

	//vai retonrar uma posiçãpo aleatoria que representa a lista de backends cadastrados.
	return backend[rand.Intn(len(backend))], true
}

func transformRespBody(body []byte) []byte {
	snakeCaseTransform := bytes.ReplaceAll(body, []byte("userId"), []byte("user_id"))
	return snakeCaseTransform
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend, exists := rp.selectBackend(r.URL.Path)

	if !exists {
		http.Error(w, "No backend found", http.StatusBadGateway)
		return
	}

	remote, err := url.Parse(backend)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	proxyReq, err := http.NewRequest(r.Method, remote.String()+r.URL.Path, r.Body)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	proxyReq.Header = r.Header

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	body := transformRespBody(response)

	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	rp := NewReverseProxy()
	http.Handle("/", rp)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
