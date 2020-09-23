package zui

import (
	"net/http"

	"github.com/torlangballe/zutil/zrpc"
)

type GetDoc struct {
	Credentials string
	Path        string
}

type GotGetDoc struct {
	Editable bool
	HTML     string
}

type DocCalls zrpc.CallsBase

var Calls = new(DocCalls)

func (c *DocCalls) GetDocument(req *http.Request, get *GetDoc, got *GotGetDoc) error {
	return nil
}

type PutDoc struct {
	Credentials string
	Path        string
	HTML        string
}
