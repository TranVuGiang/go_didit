package godidit

import (
	"io"
	"net/http"
)

type request struct {
	Method   string
	Endpoint string
	Params   any
	Header   http.Header
	Body     io.Reader
	FullURL  string
}
