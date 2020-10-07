package common

type closer interface {
	Close()
}

type urler interface {
	URL() string
}

type FakeServer interface {
	closer
	urler
}
