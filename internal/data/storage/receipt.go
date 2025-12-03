package storage

type callbackReceipt[T Record] struct {
	onConfirm func(record T, id RecordId) error
	onCancel  func(record T, err error)
	record    T
}

func newCallbackReceipt[T Record](
	record T,
	onConfirm func(record T, id RecordId) error,
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

func (d *callbackReceipt[T]) Confirm(id RecordId) error {
	return d.onConfirm(d.record, id)
}
