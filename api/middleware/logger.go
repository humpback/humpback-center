package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logger(inner http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		inner.ServeHTTP(w, r)
		log.Printf("HTTP %s\t%s\t%s\n", r.Method, r.RequestURI, time.Since(t))
	})
}
