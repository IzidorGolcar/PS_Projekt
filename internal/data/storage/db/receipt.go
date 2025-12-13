package db

type Receipt interface {
	Confirm() error
	Cancel(err error)
}
