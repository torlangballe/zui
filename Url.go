package zui

import (
	"net/url"
	"path"

	"github.com/torlangballe/zutil/ustr"
)

//  Created by Tor Langballe on /30/10/15.

type URL struct {
	neturl url.URL `json:"url"`
}

func URLNewFromstring(str string) (URL, error) {
	nu, err := url.Parse(str)
	if err == nil {
		return URL{*nu}, nil
	}
	return URL{url.URL{}}, err
}

func URLNewFromNative(nu url.URL) URL {
	return URL{nu}
}

func (u URL) OpenInBrowser(inApp, notInAppForSocial bool) {
}

func (u URL) Scheme() string {
	return u.neturl.Scheme
}

func (u URL) Host() string {
	return u.neturl.Host
}

func (u URL) Port() string {
	return u.neturl.Port()
}

func (u URL) AbsString() string {
	return u.neturl.String()
}

func (u URL) ResourcePath() string {
	return u.neturl.Path
}

func (u URL) Extension() string {
	return path.Ext(u.neturl.Path)
}

func (u URL) Anchor() string {
	return u.neturl.Fragment
}

func (u URL) Parameters() map[string]string {
	m := map[string]string{}
	for k, v := range u.neturl.Query() {
		m[k] = v[0]
	}
	return m
}

func (u URL) MultiParameters() map[string][]string {
	return u.neturl.Query()
}

func GetParametersFromArgString(parameters string) map[string]string {
	m := ustr.GetParametersFromArgString(parameters, ",", "=")
	for k, v := range m {
		m[k], _ = url.QueryUnescape(v)
	}
	return m
}
