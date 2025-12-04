package storage

import (
	"hash/fnv"
	"reflect"
)

type Record interface {
	SetId(id uint64)
	Id() uint64
}

type baseRecord struct {
	id uint64
}

func (b *baseRecord) SetId(id uint64) {
	b.id = id
}

func (b *baseRecord) Id() uint64 {
	return b.id
}

func primaryKeyHash(r Record) uint64 {
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

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("storage")
		if tag == "pk" {
			fv := v.Field(i)
			switch fv.Kind() {
			case reflect.String:
				h.Write([]byte(fv.String()))
			case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
				var b [8]byte
				x := fv.Uint()
				for i := range b {
					b[i] = byte(x >> (8 * i))
				}
				h.Write(b[:])
			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				var b [8]byte
				x := uint64(fv.Int())
				for i := range b {
					b[i] = byte(x >> (8 * i))
				}
				h.Write(b[:])
			default:
				panic("unsupported pk field type")
			}
		}
	}
	return h.Sum64()
}
