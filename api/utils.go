package api

import "net/http"

func httpError(w http.ResponseWriter, err string, code int) {
	http.Error(w, err, code)
}

func writeCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS, HEAD")
}
