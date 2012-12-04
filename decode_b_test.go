package etf

import (
	"github.com/ftrvxmtrx/testingo"
	"testing"
)

func Benchmark_DecodeStruct(b0 *testing.B) {
	b := testingo.B(b0)

	type s1 struct {
		Atom   Atom
		Uint8  uint8
		Uint16 uint16
		Uint32 uint32
		Byte   byte
		Int    int
		List   []s1
		Binary [5]byte
	}

	data := []byte{
		131, 104, 8, 100, 0, 12, 116, 104, 105, 115, 32, 105, 115, 32, 97, 116, 111, 109, 97, 255, 98, 0,
		0, 255, 255, 110, 4, 0, 255, 255, 255, 255, 97, 128, 98, 255, 255, 253, 102, 106, 107, 0, 5,
		1, 2, 3, 4, 5,
	}

	b.Logf("testing on binary of %d bytes", len(data))

	var v s1
	size, err := Decode(data, &v)
	b.AssertEq(nil, err)
	b.AssertEq(uint(len(data)), size)
	b.Log(v)

	for i := 0; i < *b.N; i++ {
		size, err = Decode(data, &v)
	}
}
