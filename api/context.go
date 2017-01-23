package api

import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/repository"
import "github.com/humpback/gounits/system"

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type (
	store map[string]interface{}

	Response struct {
		writer http.ResponseWriter
		status int
		size   int64
	}

	Context struct {
		context.Context
		ID              string
		request         *http.Request
		response        *Response
		query           url.Values
		store           store
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
	if err == nil {
		r.size += int64(n)
	}
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
		ID:              system.MakeKey(true),
		request:         r,
		response:        NewResponse(w),
		store:           make(store),
		Cluster:         cluster,
		RepositoryCache: repositorycache,
	}
}

func (c *Context) Request() *http.Request {

	return c.request
}

func (c *Context) Response() *Response {

	return c.response
}

func (c *Context) Get(key string) interface{} {

	return c.store[key]
}

func (c *Context) Set(key string, v interface{}) {

	if c.store == nil {
		c.store = make(store)
	}
	c.store[key] = v
}

func (c *Context) WriteHeader(code int) {

	c.response.WriteHeader(code)
}

func (c *Context) Query(name string) string {

	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *Context) Form(name string) string {

	return c.request.FormValue(name)
}

func (c *Context) JSON(code int, v interface{}) error {

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.response.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.response.WriteHeader(code)
	if _, err := c.response.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *Context) JSONP(code int, callback string, v interface{}) error {

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.response.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	c.response.WriteHeader(code)
	data := []byte(callback + "(")
	data = append(data, b...)
	data = append(data, []byte(");")...)
	if _, err := c.response.Write(data); err != nil {
		return err
	}
	return nil
}
