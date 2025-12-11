package relations

import "seminarska/internal/data/storage/entities"

type Receipt interface {
	Confirm() error
	Cancel(err error)
}

type callbackReceipt[T entities.Entity] struct {
	onConfirm func(record T) error
	onCancel  func(record T, err error)
	record    T
}

func newCallbackReceipt[T entities.Entity](
	record T,
	onConfirm func(record T) error,
	onCancel func(record T, err error),
) *callbackReceipt[T] {
	return &callbackReceipt[T]{
		record:    record,
		onConfirm: onConfirm,
		onCancel:  onCancel,
	}
}

func (d *callbackReceipt[T]) Cancel(err error) {
	d.onCancel(d.record, err)
}

func (d *callbackReceipt[T]) Confirm() error {
	return d.onConfirm(d.record)
}
