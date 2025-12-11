package keys

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"reflect"
)

func structHash(r any, fields []string) uint64 {
	indexable := make(map[string]bool)
	for _, f := range fields {
		indexable[f] = true
	}

	v := reflect.ValueOf(r)
	t := reflect.TypeOf(r)

	if t.Kind() == reflect.Pointer {
		v = v.Elem()
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic("not a struct")
	}

	h := fnv.New64a()
	stringerType := reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	writeBytes := func(b []byte) {
		_, _ = h.Write(b)
	}

	writeUint := func(x uint64) {
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], x)
		writeBytes(buf[:])
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !indexable[field.Name] {
			continue
		}

		fv := v.Field(i)
		ft := fv.Type()

		if ft.Implements(stringerType) {
			writeBytes([]byte(fv.Interface().(fmt.Stringer).String()))
			continue
		}
		if fv.CanAddr() && reflect.PointerTo(ft).Implements(stringerType) {
			writeBytes([]byte(fv.Addr().Interface().(fmt.Stringer).String()))
			continue
		}

		switch fv.Kind() {
		case reflect.String:
			writeBytes([]byte(fv.String()))

		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			writeUint(fv.Uint())

		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			writeUint(uint64(fv.Int()))

		default:
			panic("unsupported pk field type: " + ft.String())
		}
	}

	return h.Sum64()
}
