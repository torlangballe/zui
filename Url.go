package zgo

import (
	"net/url"
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
	return u.neturl.Scheme()
}

func (u URL) Host() string {
	return u.neturl.Host()
}

func (u URL) Port() string {
	return u.neturl.Port()
}

func (u URL) AbsString() string {
	return u.neturl.String()
}

func (u URL) ResourcePath() string {
	return u.neturl.Path()
}

func (u URL) Extension() string {
	return u.neturl.Ext()
}

func (u URL) Anchor() string {
	return u.neturl.Anchor()
}

func (u URL) Parameters() map[string]string {
	return u.neturl.Values()
}

func URLParametersFromString(parameters string) map[string]string {
	m := zstr.GetParametersFromArgString(parameters, ",", "=")
	for k, v := range m {
		m[k] = StrUrlDecode(v)
	}
	return m
}
