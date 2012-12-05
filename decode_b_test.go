package etf

import (
	"github.com/goerlang/etf/types"
	"testing"
)

func Benchmark_DecodeStruct(b *testing.B) {
	type s1 struct {
		Atom   types.ErlAtom
		priv0  int
		Uint8  uint8
		Uint16 uint16
		Uint32 uint32
		priv1  string
		Byte   byte
		Int    int
		priv2  *s1
		List   []s1
		Binary []byte
	}

	data := []byte{
		131, 104, 8, 100, 0, 12, 116, 104, 105, 115, 32, 105, 115, 32, 97, 116, 111, 109, 97, 255, 98, 0,
		0, 255, 255, 110, 4, 0, 255, 255, 255, 255, 97, 128, 98, 255, 255, 253, 102, 106, 107, 0, 5,
		1, 2, 3, 4, 5,
	}

	var v s1
	size, err := Decode(data, &v)
	if err != nil {
		b.Fatal(err)
	}
	if len(data) != size {
		b.Fatalf("expected size %d, got %d", len(data), size)
	}

	for i := 0; i < b.N; i++ {
		size, err = Decode(data, &v)
	}
}
