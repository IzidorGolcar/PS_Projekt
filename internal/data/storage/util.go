package storage

import (
	"fmt"
	"hash/fnv"
)

func stringHash(s ...any) uint64 {
	joined := fmt.Sprint(s...)
	h := fnv.New64a()
	_, err := h.Write([]byte(joined))
	if err != nil {
		panic(err)
	}
	return h.Sum64()
}
