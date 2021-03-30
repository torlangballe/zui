// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/zwords"
)

type Authentication struct {
	Email      string
	Password   string
	IsRegister bool
}

func AuthenticationOpenDialog(canCancel bool, email string, got func(auth Authentication)) {
	const column = 120.0
	v1 := StackViewVert("auth")

	emailField := TextViewNew(email, TextViewStyle{}, 20, 1)
	style := TextViewStyle{KeyboardType: KeyboardTypePassword}
	passwordField := TextViewNew("", style, 20, 1)
	register := ButtonViewNewSimple(zwords.Register(), "")
	login := ButtonViewNewSimple("Register", "")

	_, s1, _ := Labelize(emailField, "Email", column)
	v1.Add(s1, zgeo.TopLeft|zgeo.HorExpand)

	_, s2, _ := Labelize(passwordField, "Password", column)
	v1.Add(s2, zgeo.TopLeft|zgeo.HorExpand)

	h1 := StackViewHor("buttons")
	v1.Add(h1, zgeo.TopLeft|zgeo.HorExpand)

	h1.Add(register, zgeo.CenterRight)
	h1.Add(login, zgeo.CenterRight)

	register.SetPressedHandler(func() {
		var a Authentication

		a.IsRegister = true
		a.Email = emailField.Text()
		a.Password = passwordField.Text()
		go doAuth(v1, a)
	})
	login.SetPressedHandler(func() {
		var a Authentication

		a.IsRegister = false
		a.Email = emailField.Text()
		a.Password = passwordField.Text()
		go doAuth(v1, a)
	})
	if canCancel {
		cancel := ButtonViewNewSimple("Cancel", "")
		h1.Add(cancel, zgeo.CenterLeft)
		cancel.SetPressedHandler(func() {
			PresentViewClose(v1, true, nil)
		})
	}
	att := PresentViewAttributesNew()
	att.Modal = true
	PresentView(v1, att, nil, nil)
}

func doAuth(view View, a Authentication) {
	var token string
	if !zstr.IsValidEmail(a.Email) {
		AlertShow("Invalid email format:\n", a.Email)
		return
	}
	err := zrpc.ToServerClient.CallRemote("UsersCalls.Authenticate", &a, &token)
	if err != nil {
		AlertShowError("Authenticate Call Error", err)
		return
	}
	PresentViewClose(view, true, nil)
}
