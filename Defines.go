package zui

//  Defines.go
//
//  Created by Tor Langballe on /5/11/15.
//

type Object int
type AnyObject int

type UIStringer interface {
	ZUIString() string
}

func DefinesIsRunningInSimulator() bool {
	return false
}

func DefinesIsIOS() bool {
	return false
}

func DefinesIsApple() bool {
	return false
}

func DefinesIsTVBox() bool {
	return false
}

func IsTVBox() bool {
	return false
}

