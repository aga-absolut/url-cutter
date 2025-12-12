package main

import (
	"io"
	"math/rand/v2"
	"net/http"
)

var Symbols = []rune("qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM")

var Storage = make(map[string]string)

func Generate() string {
	res := make([]rune, 8)
	for i := range res {
		res[i] = Symbols[rand.IntN(len(Symbols))]
	}
	return string(res)
}
func ShortUrl(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		bodyByte, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		shortUrl := Generate()
		Storage[shortUrl] = string(bodyByte)

		w.Header().Set("Content-Type: ", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/" + shortUrl))
	case http.MethodGet:
		path := r.URL.Path
		if path == "/" {
			http.Error(w, "Bad Request", 400)
			return
		}

		shortPath := path[1:]
		logUrl := Storage[shortPath]

		if logUrl != "" {
			w.Header().Set("Location", "")
			w.WriteHeader(http.StatusTemporaryRedirect)
			w.Write([]byte(logUrl))
		} else {
			http.Error(w, "Not found", 400)
		}

	default:
		http.Error(w, "Bad Request", 400)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, ShortUrl)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
