package zgo

func PresentViewURL(surl string) {
	WindowJS.Get("location").Call("assign", surl)
}

// https://bubblin.io/blog/fullscreen-api-ipad
// https://medium.com/@firt/iphone-11-ipados-and-ios-13-for-pwas-and-web-development-5d5d9071cc49