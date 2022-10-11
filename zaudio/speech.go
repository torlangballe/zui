package zaudio

type Voice interface {
	Name() string
	Female() bool
	Language() string // returns
}
