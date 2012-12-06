package etf

import (
	"bytes"
	"github.com/goerlang/etf/types"
	"math/big"
	"testing"
)

func TestDecodeArray(t *testing.T) {
	var v [3]byte

	in := bytes.NewBuffer([]byte{131, 107, 0, 3, 1, 2, 3})
	if err := Decode(in, &v); err != nil {
		t.Fatal(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := []byte{1, 2, 3}; bytes.Compare(v[:], exp) != 0 {
		t.Errorf("expected %v, got %v", exp, v)
	}
}

func TestDecodeBigInt(t *testing.T) {
	var v *big.Int

	in := bytes.NewBuffer([]byte{
		131, 110, 15, 0, 0, 0, 0, 0, 16, 159,
		75, 179, 21, 7, 201, 123, 206, 151, 192,
	})
	exp, _ := new(big.Int).SetString("1000000000000000000000000000000000000", 10)
	if err := Decode(in, &v); err != nil {
		t.Fatal(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp.Cmp(v) != 0 {
		t.Errorf("expected %v, got %v", exp, v)
	}
}

func TestDecodeBinary(t *testing.T) {
	var data []byte

	in := bytes.NewBuffer([]byte{131, 109, 0, 0, 0, 3, 1, 2, 3})
	err := Decode(in, &data)
	if err != nil {
		t.Fatal(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := []byte{1, 2, 3}; bytes.Compare(data, exp) != 0 {
		t.Errorf("expected %v, got %v", exp, data)
	}
}

func TestDecodeString(t *testing.T) {
	var s string

	in := bytes.NewBuffer([]byte{131, 107, 0, 3, 49, 50, 51})
	err := Decode(in, &s)
	if err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := "123"; s != exp {
		t.Errorf("expected %v, got %v", exp, s)
	}
}

func TestDecodeStruct(t *testing.T) {
	type testStruct struct {
		types.ErlAtom
		X uint8
		S string
	}
	var ts testStruct

	in := bytes.NewBuffer([]byte{
		131, 104, 3, 100, 0, 4, 98, 108, 97, 104, 97, 4, 108, 0, 0, 0, 4, 98,
		0, 0, 4, 68, 98, 0, 0, 4, 75, 98, 0, 0, 4, 50, 98, 0, 0, 4, 48, 106,
	})
	if err := Decode(in, &ts); err != nil {
		t.Fatal(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := (testStruct{types.ErlAtom("blah"), 4, "фыва"}); ts != exp {
		t.Errorf("expected %v, got %v", exp, ts)
	}
}

func TestDecodeStruct2(t *testing.T) {
	type testStruct2 struct {
		j string
		F float32
		X int
		i [2]byte
		S string
	}
	var ts testStruct2

	in := bytes.NewBuffer([]byte{
		131, 104, 3, 99, 50, 46, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57,
		57, 57, 57, 56, 56, 56, 57, 56, 101, 45, 48, 49, 0, 0, 0, 0, 0, 97, 4, 108,
		0, 0, 0, 4, 98, 0, 0, 4, 68, 98, 0, 0, 4, 75, 98, 0, 0, 4, 50, 98, 0, 0, 4,
		48, 106,
	})
	if err := Decode(in, &ts); err != nil {
		t.Fatal(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := (testStruct2{"", 0.3, 4, [2]byte{0, 0}, "фыва"}); ts != exp {
		t.Errorf("expected %v, got %v", exp, ts)
	}
}

func TestDecodeStruct3(t *testing.T) {
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

	in := bytes.NewBuffer([]byte{
		131,
		104, 2,
		104, 3,
		100, 0, 4, 98, 108, 97,
		104, 97, 4,
		108, 0, 0, 0, 4,
		98, 0, 0, 4, 68,
		98, 0, 0, 4, 75,
		98, 0, 0, 4, 50,
		98, 0, 0, 4, 48,
		106,
		98, 0, 0, 2, 154,
	})
	exp := testStruct3{
		testStruct{
			types.ErlAtom("blah"),
			4,
			nilBig,
			"фыва",
		},
		nilArr,
		666,
	}
	if err := Decode(in, &ts); err != nil {
		t.Fatal(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if ts != exp {
		t.Errorf("expected %v, got %v", exp, ts)
	}
}
