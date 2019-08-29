package zgo

import "errors"

//  Created by Tor Langballe on /5/11/15.

type Error struct {
	Message string
	Code    int
	Domain  string
}

func ErrorNew(message string, code int, domain string) *Error {
	return &Error{message, code, domain}
}

func (e *Error) Error() error {
	return errors.New(e.Message)
}

func (e *Error) GetMessage() string {
	return e.Message
}

func ErrorFromErr(err error) *Error {
	if err == nil {
		return nil
	}
	return ErrorNew(err.Error(), 0, "")
}

var ErrorGeneral = ErrorNew("Zed", 1, "Zetrus")
var ErrorUrlDomain = "url"
