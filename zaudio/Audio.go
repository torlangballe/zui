package zaudio

type Audio struct {
	audioNative
}

const ResultMessageHeaderKey = "X-ZAudio-Rec-Result"

func Play(url string) {
	a := New(url)
	a.Play(nil)
}
