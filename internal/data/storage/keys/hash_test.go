package keys

import (
	"testing"
)

type TestStruct struct {
	Field1 string `storage:"pk"`
	Field2 int64
}

func TestStructHash(t *testing.T) {
	type args struct {
		e1, e2 any
		fields []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"structHash 1",
			args{
				e1:     TestStruct{"A", 0},
				e2:     TestStruct{"A", 1},
				fields: []string{"Field1"}},
			true,
		},
		{
			"structHash 2",
			args{
				e1:     TestStruct{"A", 0},
				e2:     TestStruct{"B", 0},
				fields: []string{"Field1"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := structHash(tt.args.e1, tt.args.fields)
			h2 := structHash(tt.args.e2, tt.args.fields)

			if got := h1 == h2; got != tt.want {
				t.Errorf("structHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
