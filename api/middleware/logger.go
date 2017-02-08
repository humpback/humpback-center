package middleware

import "github.com/humpback/gounits/logger"

import (
	"net/http"
	"time"
)

func Logger(inner http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		inner.ServeHTTP(w, r)
		logger.INFO("[#api#] HTTP %s\t%s\t%s", r.Method, r.RequestURI, time.Since(t))
	})
}
