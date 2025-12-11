package storage

import (
	"seminarska/proto/datalink"
	"testing"
)

type testStruct struct {
	baseRecord
	name string `storage:"pk"`
	age  int
}

func (t *testStruct) ToDatalinkRecord() *datalink.Record {
	return nil
}

func TestBaseRecord(t *testing.T) {
	type args struct {
		record *testStruct
		autoId int64
	}
	type testCase struct {
		name string
		args args
		want int64
	}
	tests := []testCase{
		{
			name: "TestBaseRecord",
			args: args{
				&testStruct{name: "A", age: 1},
				1,
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.record.SetId(tt.args.autoId)
			if got := tt.args.record.Id(); got != tt.want {
				t.Errorf("BaseRecord.Id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_primaryKeyHash(t *testing.T) {
	type args struct {
		a *testStruct
		b *testStruct
	}
	type testCase struct {
		name string
		args args
		want bool
	}
	tests := []testCase{
		{
			name: "different uniqueIndex",
			args: args{
				&testStruct{name: "A", age: 1},
				&testStruct{name: "B", age: 1},
			},
			want: false,
		},
		{
			name: "identical uniqueIndex",
			args: args{
				&testStruct{name: "A", age: 1},
				&testStruct{name: "A", age: 2},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashA := primaryKeyHash(tt.args.a)
			hashB := primaryKeyHash(tt.args.b)
			if got := hashA == hashB; got != tt.want {
				t.Errorf("%v == %v, want %v", hashA, hashB, tt.want)
			}
		})
	}
}
