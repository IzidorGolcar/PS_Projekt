package chain

type messageCounter struct {
	value int64
}

func newMessageCounter(initial int64) *messageCounter {
	return &messageCounter{value: initial}
}

func (c *messageCounter) next() int64 {
	c.value++
	return c.value
}
