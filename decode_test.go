package etf

import (
	"bytes"
	"github.com/goerlang/etf/types"
	"math/big"
	"testing"
)

func Test_decodeArray(t *testing.T) {
	var v [3]byte

	in := []byte{131, 107, 0, 3, 1, 2, 3}
	if size, err := Decode(in, &v); err != nil {
		t.Fatal(err)
	} else if size != len(in) {
		t.Errorf("expected size %d, got %d", len(in), size)
	} else if exp := []byte{1, 2, 3}; bytes.Compare(v[:], exp) != 0 {
		t.Errorf("expected %v, got %v", exp, v)
	}
}

func Test_decodeBigInt(t *testing.T) {
	var v *big.Int

	in := []byte{131, 110, 15, 0, 0, 0, 0, 0, 16, 159, 75, 179, 21, 7, 201, 123, 206, 151, 192}
	exp, _ := new(big.Int).SetString("1000000000000000000000000000000000000", 10)
	if size, err := Decode(in, &v); err != nil {
		t.Fatal(err)
	} else if size != len(in) {
		t.Errorf("expected size %d, got %d", len(in), size)
	} else if exp.Cmp(v) != 0 {
		t.Errorf("expected %v, got %v", exp, v)
	}
}

func Test_decodeBinary(t *testing.T) {
	var data []byte

	in := []byte{131, 109, 0, 0, 0, 3, 1, 2, 3}
	size, err := Decode(in, &data)
	if err != nil {
		t.Fatal(err)
	} else if size != len(in) {
		t.Errorf("expected size %d, got %d", len(in), size)
	} else if exp := []byte{1, 2, 3}; bytes.Compare(data, exp) != 0 {
		t.Errorf("expected %v, got %v", exp, data)
	}
}

func Test_decodeString(t *testing.T) {
	var s string

	in := []byte{131, 107, 0, 3, 49, 50, 51}
	size, err := Decode(in, &s)
	if err != nil {
		t.Error(err)
	}

	exp := "123"
	if size != len(in) {
		t.Errorf("expected size %d, got %d", len(in), size)
	} else if s != exp {
		t.Errorf("expected %v, got %v", exp, s)
	}
}

func Test_decodeStruct(t *testing.T) {
	type testStruct struct {
		types.ErlAtom
		X uint8
		S string
	}
	var ts testStruct

	in := []byte{
		131, 104, 3, 100, 0, 4, 98, 108, 97, 104, 97, 4, 108, 0, 0, 0, 4, 98,
		0, 0, 4, 68, 98, 0, 0, 4, 75, 98, 0, 0, 4, 50, 98, 0, 0, 4, 48, 106,
	}
	if size, err := Decode(in, &ts); err != nil {
		t.Fatal(err)
	} else if size != len(in) {
		t.Errorf("expected size %d, got %d", len(in), size)
	} else if exp := (testStruct{types.ErlAtom("blah"), 4, "фыва"}); ts != exp {
		t.Errorf("expected %v, got %v", exp, ts)
	}
}

func Test_decodeStruct2(t *testing.T) {
	type testStruct2 struct {
		j string
		F float32
		X int
		i [2]byte
		S string
	}
	var ts testStruct2

	in := []byte{
		131, 104, 3, 99, 50, 46, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57,
		57, 57, 56, 56, 56, 57, 56, 101, 45, 48, 49, 0, 0, 0, 0, 0, 97, 4, 108, 0, 0,
		0, 4, 98, 0, 0, 4, 68, 98, 0, 0, 4, 75, 98, 0, 0, 4, 50, 98, 0, 0, 4, 48, 106,
	}
	if size, err := Decode(in, &ts); err != nil {
		t.Fatal(err)
	} else if size != len(in) {
		t.Errorf("expected size %d, got %d", len(in), size)
	} else if exp := (testStruct2{"", 0.3, 4, [2]byte{0, 0}, "фыва"}); ts != exp {
		t.Errorf("expected %v, got %v", exp, ts)
	}
}

func Test_decodeStruct3(t *testing.T) {
	type testStruct struct {
		types.ErlAtom
		X uint8
		i *big.Int
		S string
	}
	type testStruct3 struct {
		T testStruct
		i [2]byte
		Y int
	}
	var ts testStruct3

	nilBig := (*big.Int)(nil)
	nilArr := [2]byte{0, 0}

	in := []byte{
		131, 104, 2, 104, 3, 100, 0, 4, 98, 108, 97, 104, 97, 4, 108, 0, 0, 0, 4, 98,
		0, 0, 4, 68, 98, 0, 0, 4, 75, 98, 0, 0, 4, 50, 98, 0, 0, 4, 48, 106, 98, 0, 0, 2, 154,
	}
	if size, err := Decode(in, &ts); err != nil {
		t.Fatal(err)
	} else if size != len(in) {
		t.Errorf("expected size %d, got %d", len(in), size)
	} else if exp := (testStruct3{testStruct{types.ErlAtom("blah"), 4, nilBig, "фыва"}, nilArr, 666}); ts != exp {
		t.Errorf("expected %v, got %v", exp, ts)
	}
}
