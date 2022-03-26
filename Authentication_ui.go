//go:build zui
// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/zusers"
	"github.com/torlangballe/zutil/zwords"
)

const emailKey = "zui.AuthenticationEmail"

var (
	AuthenticationCurrentUserID int64
	AuthenticationCurrentToken  string
)

func AuthenticationOpenDialog(canCancel bool, got func(auth zusers.AuthenticationResult)) {
	const column = 120.0
	v1 := StackViewVert("auth")
	v1.SetSpacing(10)
	v1.SetMarginS(zgeo.Size{10, 10})
	v1.SetBGColor(zgeo.ColorNewGray(0.9, 1))
	email, _ := zkeyvalue.DefaultStore.GetString(emailKey)
	style := TextViewStyle{KeyboardType: KeyboardTypeEmailAddress}
	emailField := TextViewNew(email, style, 20, 1)
	style = TextViewStyle{KeyboardType: KeyboardTypePassword}
	passwordField := TextViewNew("", style, 20, 1)
	register := ButtonNew(zwords.Register())
	register.SetMinWidth(90)
	login := ButtonNew(zwords.Login())
	login.SetMinWidth(90)
	login.MakeEnterDefault()

	_, s1, _ := Labelize(emailField, "Email", column, zgeo.CenterLeft)
	v1.Add(s1, zgeo.TopLeft|zgeo.HorExpand)

	_, s2, _ := Labelize(passwordField, "Password", column, zgeo.CenterLeft)
	v1.Add(s2, zgeo.TopLeft|zgeo.HorExpand)

	h1 := StackViewHor("buttons")
	v1.Add(h1, zgeo.TopLeft|zgeo.HorExpand, zgeo.Size{0, 14})

	h1.Add(register, zgeo.CenterRight)
	h1.Add(login, zgeo.CenterRight)

	register.SetPressedHandler(func() {
		var a zusers.Authentication

		a.IsRegister = true
		a.Email = emailField.Text()
		a.Password = passwordField.Text()
		go doAuth(v1, a, got)
	})
	login.SetPressedHandler(func() {
		var a zusers.Authentication

		a.IsRegister = false
		a.Email = emailField.Text()
		a.Password = passwordField.Text()
		go doAuth(v1, a, got)
	})
	if canCancel {
		cancel := ImageButtonViewNewSimple("Cancel", "")
		h1.Add(cancel, zgeo.CenterLeft)
		cancel.SetPressedHandler(func() {
			PresentViewClose(v1, true, nil)
		})
	}
	att := PresentViewAttributesNew()
	att.Modal = true
	PresentView(v1, att, nil, nil)
}

func doAuth(view View, a zusers.Authentication, got func(auth zusers.AuthenticationResult)) {
	var aret zusers.AuthenticationResult
	if !zstr.IsValidEmail(a.Email) {
		AlertShow("Invalid email format:\n", a.Email)
		return
	}
	zkeyvalue.DefaultStore.SetString(a.Email, emailKey, true)

	err := zrpc.ToServerClient.CallRemote("UsersCalls.Authenticate", &a, &aret)
	if err != nil {
		AlertShowError(err, "Authenticate Call Error")
		return
	}
	zlog.Info("Do auth:", aret)
	PresentViewClose(view, false, func(dismissed bool) {
		zlog.Info("Do auth2:", dismissed)
		if !dismissed {
			got(aret)
		}
	})
}

func CheckAndDoAuthentication(client *zrpc.Client, canCancel bool, got func(auth zusers.AuthenticationResult)) {
	const tokenKey = "zui.AuthenticationToken"
	var user zusers.User

	client.ID, _ = zkeyvalue.DefaultStore.GetString(tokenKey)
	// zlog.Info("CheckAndDoAuthentication:", client.ID)
	if zrpc.ToServerClient.ID != "" {
		err := client.CallRemote("UsersCalls.CheckIfUserLoggedInWithZRPCHeaderToken", nil, &user)
		if err == nil {
			var auth zusers.AuthenticationResult
			auth.ID = user.ID
			AuthenticationCurrentUserID = user.ID
			auth.Token = client.ID
			AuthenticationCurrentToken = client.ID
			// zlog.Info("CheckAndDoAuthentication existed:", auth)
			if got != nil {
				got(auth)
			}
			return
		}
		AlertShowError(err, "Authentication Error")
	}
	AuthenticationOpenDialog(canCancel, func(auth zusers.AuthenticationResult) {
		client.ID = auth.Token
		zkeyvalue.DefaultStore.SetString(auth.Token, tokenKey, true)
		zlog.Info("got auth:", auth)
		if got != nil {
			AuthenticationCurrentUserID = auth.ID
			AuthenticationCurrentToken = auth.Token
			got(auth)
		}
	})
}
