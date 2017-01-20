package api

import "github.com/humpback/humpback-center/cluster"

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
)

type Context struct {
	context.Context
	Cluster   *cluster.Cluster
	tlsConfig *tls.Config
}

func NewContext(cluster *cluster.Cluster, tlsConfig *tls.Config) *Context {

	return &Context{
		Cluster:   cluster,
		tlsConfig: tlsConfig,
	}
}

func (ctx *Context) JSON(w http.ResponseWriter, code int, v interface{}) error {

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}

func (ctx *Context) JSONP(w http.ResponseWriter, code int, callback string, v interface{}) error {

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(code)
	data := []byte(callback + "(")
	data = append(data, b...)
	data = append(data, []byte(");")...)
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}
