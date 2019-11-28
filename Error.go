package zgo

import (
	"errors"
	"fmt"
	"strings"
)

//  Created by Tor Langballe on /5/11/15.

type CodeDomainError struct {
	Message string
	Code    int
	Domain  string
}

func (e *CodeDomainError) Error() error {
	return errors.New(e.Message)
}

func ErrorNew(parts ...interface{}) error {
	p := strings.TrimSpace(fmt.Sprintln(parts...))
	return errors.New(p)
}

// var ErrorGeneral = ErrorNew("Zed", 1, "Zetrus")
// var ErrorUrlDomain = "url"
