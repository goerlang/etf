package parse

import (
	"bytes"
	. "github.com/goerlang/etf/types"
	"math/big"
	"testing"
)

func TestAtom(t *testing.T) {
	// 'abc'
	in := bytes.NewBuffer([]byte{100, 0, 3, 97, 98, 99})
	if v, err := Atom(in); err != nil {
		t.Fatal(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := ErlAtom("abc"); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// ''
	in = bytes.NewBuffer([]byte{100, 0, 0})
	if v, err := Atom(in); err != nil {
		t.Fatal(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := ErlAtom(""); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 'abc' as SmallAtom
	in = bytes.NewBuffer([]byte{115, 3, 97, 98, 99})
	if v, err := Atom(in); err != nil {
		t.Fatal(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := ErlAtom("abc"); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// '' as SmallAtom
	in = bytes.NewBuffer([]byte{115, 0})
	if v, err := Atom(in); err != nil {
		t.Fatal(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := ErlAtom(""); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// error (ends abruptly)
	if _, err := Atom(bytes.NewBuffer([]byte{100, 0, 4, 97, 98, 99})); err == nil {
		t.Error("err == nil")
	}

	// error (bad length)
	if _, err := Atom(bytes.NewBuffer([]byte{100})); err == nil {
		t.Error("err == nil")
	}
}

func TestBigInt(t *testing.T) {
	// (1<<2040)
	in := bytes.NewBuffer([]byte{
		111, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	})
	if v, err := BigInt(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if v, exp := new(big.Int).Lsh(big.NewInt(1), 2040).Cmp(v), 0; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// -(1<<2040)
	in = bytes.NewBuffer([]byte{
		111, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	})
	exp := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 2040))
	if v, err := BigInt(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp.Cmp(v) != 0 {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 0 (small big)
	in = bytes.NewBuffer([]byte{110, 0, 0})
	if v, err := BigInt(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if v, exp := big.NewInt(0).Cmp(v), 0; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 0 (large big)
	in = bytes.NewBuffer([]byte{111, 0, 0, 0, 0, 0})
	if v, err := BigInt(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if v, exp := big.NewInt(0).Cmp(v), 0; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}
}

func TestBool(t *testing.T) {
	// true
	in := bytes.NewBuffer([]byte{100, 0, 4, 't', 'r', 'u', 'e'})
	if v, err := Bool(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := true; exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// false
	in = bytes.NewBuffer([]byte{100, 0, 5, 'f', 'a', 'l', 's', 'e'})
	if v, err := Bool(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := false; exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// error
	in = bytes.NewBuffer([]byte{100, 0, 3, 97, 98, 99})
	if _, err := Bool(in); err == nil {
		t.Error("err == nil")
	}
}

func TestFloat64(t *testing.T) {
	// 0.1
	in := bytes.NewBuffer([]byte{
		99, 49, 46, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48,
		48, 48, 48, 53, 53, 53, 49, 101, 45, 48, 49, 0, 0, 0, 0, 0,
	})
	if v, err := Float64(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := 0.1; exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 0.1
	in = bytes.NewBuffer([]byte{70, 63, 185, 153, 153, 153, 153, 153, 154})
	if v, err := Float64(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := 0.1; exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// error (31 bytes instead of 32)
	in = bytes.NewBuffer([]byte{
		99, 49, 46, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48,
		48, 48, 48, 53, 53, 53, 49, 101, 45, 48, 49, 0, 0, 0, 0,
	})
	if _, err := Float64(in); err == nil {
		t.Error("err == nil")
	}

	// error (fail on Sscanf)
	in = bytes.NewBuffer([]byte{
		99, 99, 46, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48,
		48, 48, 48, 53, 53, 53, 49, 101, 45, 48, 49, 0, 0, 0, 0, 0,
	})
	if _, err := Float64(in); err == nil {
		t.Error("err == nil")
	}

	// error (bad length)
	if _, err := Float64(bytes.NewBuffer([]byte{99})); err == nil {
		t.Error("err == nil")
	}

	// error (bad length)
	if _, err := Float64(bytes.NewBuffer([]byte{70})); err == nil {
		t.Error("err == nil")
	}
}

func TestInt64(t *testing.T) {
	// 255
	in := bytes.NewBuffer([]byte{97, 255})
	if v, err := Int64(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := int64(255); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 0x7fffffff
	in = bytes.NewBuffer([]byte{98, 127, 255, 255, 255})
	if v, err := Int64(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := int64(0x7fffffff); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// -0x80000000
	in = bytes.NewBuffer([]byte{98, 128, 0, 0, 0})
	if v, err := Int64(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := int64(-0x80000000); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 0x7fffffffffffffff
	in = bytes.NewBuffer([]byte{110, 8, 0, 255, 255, 255, 255, 255, 255, 255, 127})
	if v, err := Int64(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := int64(0x7fffffffffffffff); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// -0x8000000000000000
	in = bytes.NewBuffer([]byte{110, 8, 1, 0, 0, 0, 0, 0, 0, 0, 128})
	if v, err := Int64(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := int64(-0x8000000000000000); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// error (bad length)
	for _, b := range []byte{97, 98, 110, 111} {
		if _, err := Int64(bytes.NewBuffer([]byte{b})); err == nil {
			t.Error("err == nil (%d)", b)
		}
	}

	// error (0x8000000000000000)
	in = bytes.NewBuffer([]byte{110, 8, 0, 0, 0, 0, 0, 0, 0, 0, 128})
	if _, err := Int64(in); err == nil {
		t.Error("err == nil")
	}

	// error (-0x8000000000000001)
	in = bytes.NewBuffer([]byte{110, 8, 1, 1, 0, 0, 0, 0, 0, 0, 128})
	if _, err := Int64(in); err == nil {
		t.Error("err == nil")
	}
}

func TestString(t *testing.T) {
	// "" (nil)
	in := bytes.NewBuffer([]byte{106})
	if v, err := String(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := ""; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "" (empty string)
	in = bytes.NewBuffer([]byte{107, 0, 0})
	if v, err := String(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := ""; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "" (empty list)
	in = bytes.NewBuffer([]byte{108, 0, 0, 0, 0, 106})
	if v, err := String(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := ""; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "" (empty binary)
	in = bytes.NewBuffer([]byte{109, 0, 0, 0, 0})
	if v, err := String(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := ""; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "abc"
	in = bytes.NewBuffer([]byte{107, 0, 3, 97, 98, 99})
	if v, err := String(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := "abc"; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// <<"abc">>
	in = bytes.NewBuffer([]byte{109, 0, 0, 0, 3, 97, 98, 99})
	if v, err := String(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := "abc"; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "фыва", where the last one is of erlSmallBig type
	in = bytes.NewBuffer([]byte{
		108, 0, 0, 0, 4, 98, 0, 0, 4, 68, 98, 0, 0, 4,
		75, 98, 0, 0, 4, 50, 110, 2, 0, 48, 4, 106,
	})
	if v, err := String(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := "фыва"; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "фыва", where the last one is of erlLargeBig type
	in = bytes.NewBuffer([]byte{
		108, 0, 0, 0, 4, 98, 0, 0, 4, 68, 98, 0, 0, 4,
		75, 98, 0, 0, 4, 50, 111, 0, 0, 0, 2, 0, 48, 4, 106,
	})
	if v, err := String(in); err != nil {
		t.Error(err)
	} else if l := in.Len(); l != 0 {
		t.Errorf("buffer len %d", l)
	} else if exp := "фыва"; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// error (wrong length) in string
	in = bytes.NewBuffer([]byte{107, 0, 3, 97, 98})
	if _, err := String(in); err == nil {
		t.Error("err == nil")
	}

	// error (wrong length) in binary string
	in = bytes.NewBuffer([]byte{109, 0, 0, 0, 3, 97, 98})
	if _, err := String(in); err == nil {
		t.Error("err == nil")
	}

	// error (improper list) [$a,$b,$c|0]
	in = bytes.NewBuffer([]byte{108, 0, 0, 0, 3, 97, 98, 99, 0})
	if _, err := String(in); err == nil {
		t.Error("err == nil")
	}

	// error (bad length)
	for _, b := range []byte{107, 108, 109} {
		if _, err := String(bytes.NewBuffer([]byte{b})); err == nil {
			t.Error("err == nil (%d)", b)
		}
	}
}
