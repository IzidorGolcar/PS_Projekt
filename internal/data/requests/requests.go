package requests

import "fmt"

type Request interface {
	Done()
	cancel(err error)
}

type baseRequest struct {
	done chan struct{}
	err  error
}

func newBaseRequest() *baseRequest {
	return &baseRequest{done: make(chan struct{})}
}

func (b *baseRequest) Done() {
	close(b.done)
}

func (b *baseRequest) cancel(err error) {
	b.err = fmt.Errorf("canceled: %w", err)
	close(b.done)
}

type NewUser struct {
	baseRequest
	name string
}
