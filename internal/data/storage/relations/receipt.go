package relations

type Receipt interface {
	Confirm() error
	Cancel(err error)
}
