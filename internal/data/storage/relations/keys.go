package relations

import (
	"reflect"
	"seminarska/internal/data/storage/keys"
)

func newUniqueIndex[E any]() *keys.Index {
	return keys.NewIndex(getUniqueFields[E]()...)
}

func getUniqueFields[E any]() (fields []string) {
	e := reflect.TypeOf((*E)(nil)).Elem()
	if e.Kind() != reflect.Struct {
		panic("Type must be a struct")
	}
	for i := 0; i < e.NumField(); i++ {
		f := e.Field(i)
		if f.Tag.Get("db") == "unique" {
			fields = append(fields, f.Name)
		}
	}
	return
}
