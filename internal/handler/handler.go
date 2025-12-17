package handler

import (
	"io"
	"math/rand/v2"
	"net/http"

	"github.com/go-chi/chi/v5"
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
// func ShortURL(w http.ResponseWriter, r *http.Request) {
// 	switch r.Method {
// 	case http.MethodPost:
// 		bodyByte, err := io.ReadAll(r.Body)

// 		if err != nil {
// 			http.Error(w, err.Error(), 400)
// 			return
// 		}
// 		shortURL := Generate()
// 		Storage[shortURL] = string(bodyByte)

// 		w.Header().Set("Content-Type", "text/plain")
// 		w.WriteHeader(http.StatusCreated)
// 		w.Write([]byte("http://localhost:8080/" + shortURL))

// 	case http.MethodGet:
// 		path := r.URL.Path
// 		if path == "/" {
// 			http.Error(w, "Bad Request", 400)
// 			return
// 		}

// 		shortPath := path[1:]
// 		logURL := Storage[shortPath]

// 		if logURL != "" {
// 			w.Header().Set("Location", string(logURL))
// 			w.WriteHeader(http.StatusTemporaryRedirect)
// 		} else {
// 			http.Error(w, "Not found", 404)
// 		}

// 	default:
// 		http.Error(w, "Bad Request", 400)
// 	}
// }

func ShortPostReq(w http.ResponseWriter, r *http.Request) {
	resp, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	shortURL := Generate()
	Storage[shortURL] = string(resp)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + shortURL))
}

func ShortGetReq(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "rf")

	if path == "" {
		http.Error(w, "Bad request", 477)
		return
	}
	resURL := Storage[path]

	if resURL != "" {
		w.Header().Set("Location", resURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Not found", 404)
	}
}
