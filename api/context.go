package api

import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/repository"

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type (
	Response struct {
		writer http.ResponseWriter
		status int
		size   int64
	}

	Context struct {
		context.Context
		request         *http.Request
		response        *Response
		query           url.Values
		Cluster         *cluster.Cluster
		RepositoryCache *repository.RepositoryCache
	}
)

func NewResponse(w http.ResponseWriter) *Response {

	return &Response{
		writer: w,
	}
}

func (r *Response) SetWriter(w http.ResponseWriter) {

	r.writer = w
}

func (r *Response) Header() http.Header {

	return r.writer.Header()
}

func (r *Response) Writer() http.ResponseWriter {

	return r.writer
}

func (r *Response) WriteHeader(code int) {

	r.status = code
	r.writer.WriteHeader(code)
}

func (r *Response) Write(b []byte) (int, error) {

	n, err := r.writer.Write(b)
	r.size += int64(n)
	return n, err
}

func (r *Response) Flush() {

	r.writer.(http.Flusher).Flush()
}

func (r *Response) Size() int64 {

	return r.size
}

func (r *Response) Status() int {

	return r.status
}

func NewContext(w http.ResponseWriter, r *http.Request,
	cluster *cluster.Cluster, repositorycache *repository.RepositoryCache) *Context {

	return &Context{
		request:         r,
		response:        NewResponse(w),
		Cluster:         cluster,
		RepositoryCache: repositorycache,
	}
}

func (ctx *Context) Request() *http.Request {

	return ctx.request
}

func (ctx *Context) Response() *Response {

	return ctx.response
}

func (ctx *Context) WriteHeader(code int) {

	ctx.response.WriteHeader(code)
}

func (ctx *Context) Query(name string) string {

	if ctx.query == nil {
		ctx.query = ctx.request.URL.Query()
	}
	return ctx.query.Get(name)
}

func (ctx *Context) Form(name string) string {

	return ctx.request.FormValue(name)
}

func (ctx *Context) JSON(code int, v interface{}) error {

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ctx.response.Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.response.WriteHeader(code)
	if _, err := ctx.response.Write(data); err != nil {
		return err
	}
	return nil
}

func (ctx *Context) JSONP(code int, callback string, v interface{}) error {

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ctx.response.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	ctx.response.WriteHeader(code)
	data := []byte(callback + "(")
	data = append(data, b...)
	data = append(data, []byte(");")...)
	if _, err := ctx.response.Write(data); err != nil {
		return err
	}
	return nil
}
