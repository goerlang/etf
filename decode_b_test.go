package etf

import (
	"bytes"
	"github.com/goerlang/etf/types"
	"testing"
)

func BenchmarkDecodeStruct(b *testing.B) {
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
		131, 104, 8, 100, 0, 12, 116, 104, 105, 115, 32, 105, 115, 32, 97, 116, 111,
		109, 97, 255, 98, 0, 0, 255, 255, 110, 4, 0, 255, 255, 255, 255, 97, 128, 98,
		255, 255, 253, 102, 106, 107, 0, 5, 1, 2, 3, 4, 5,
	}
	in := bytes.NewBuffer(data)

	var v s1
	err := Decode(in, &v)
	if err != nil {
		b.Fatal(err)
	} else if l := in.Len(); l != 0 {
		b.Fatalf("buffer len %d", l)
	}

	for i := 0; i < b.N; i++ {
		in = bytes.NewBuffer(data)
		err = Decode(in, &v)
	}
}
