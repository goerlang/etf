package parse

import (
	. "github.com/goerlang/etf/types"
	"math/big"
	"testing"
)

func isStructuralError(err error) bool {
	switch err.(type) {
	case StructuralError:
		return true
	}
	return false
}

func Test_Atom(t *testing.T) {
	// 'abc'
	if v, size, err := Atom([]byte{100, 0, 3, 97, 98, 99}); err != nil {
		t.Fatal(err)
	} else if size != 6 {
		t.Errorf("expected size %d, got %d", 6, size)
	} else if exp := ErlAtom("abc"); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// ''
	if v, size, err := Atom([]byte{100, 0, 0}); err != nil {
		t.Fatal(err)
	} else if size != 3 {
		t.Errorf("expected size %d, got %d", 3, size)
	} else if exp := ErlAtom(""); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 'abc' as SmallAtom
	if v, size, err := Atom([]byte{115, 3, 97, 98, 99}); err != nil {
		t.Fatal(err)
	} else if size != 5 {
		t.Errorf("expected size %d, got %d", 5, size)
	} else if exp := ErlAtom("abc"); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// '' as SmallAtom
	if v, size, err := Atom([]byte{115, 0}); err != nil {
		t.Fatal(err)
	} else if size != 2 {
		t.Errorf("expected size %d, got %d", 2, size)
	} else if exp := ErlAtom(""); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// error (ends abruptly)
	if _, _, err := Atom([]byte{100, 0, 4, 97, 98, 99}); err == nil {
		t.Error("err == nil")
	} else if !isStructuralError(err) {
		t.Errorf("error is not StructuralError, but %T", err)
	}

	// error (bad length)
	if _, _, err := Atom([]byte{100}); err == nil {
		t.Error("err == nil")
	} else if !isStructuralError(err) {
		t.Errorf("error is not StructuralError, but %T", err)
	}
}

func Test_BigInt(t *testing.T) {
	// (1<<2040)
	b := []byte{
		111, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 1,
	}
	if v, size, err := BigInt(b); err != nil {
		t.Error(err)
	} else if size != len(b) {
		t.Errorf("expected size %d, got %d", len(b), size)
	} else if v, exp := new(big.Int).Lsh(big.NewInt(1), 2040).Cmp(v), 0; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// -(1<<2040)
	b = []byte{
		111, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 1,
	}
	if v, size, err := BigInt(b); err != nil {
		t.Error(err)
	} else if size != len(b) {
		t.Errorf("expected size %d, got %d", len(b), size)
	} else if v, exp := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 2040)).Cmp(v), 0; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 0 (small big)
	b = []byte{110, 0, 0}
	if v, size, err := BigInt(b); err != nil {
		t.Error(err)
	} else if size != len(b) {
		t.Errorf("expected size %d, got %d", len(b), size)
	} else if v, exp := big.NewInt(0).Cmp(v), 0; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 0 (large big)
	b = []byte{111, 0, 0, 0, 0, 0}
	if v, size, err := BigInt(b); err != nil {
		t.Error(err)
	} else if size != len(b) {
		t.Errorf("expected size %d, got %d", len(b), size)
	} else if v, exp := big.NewInt(0).Cmp(v), 0; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}
}

func Test_Bool(t *testing.T) {
	// true
	if v, size, err := Bool([]byte{100, 0, 4, 't', 'r', 'u', 'e'}); err != nil {
		t.Error(err)
	} else if size != 7 {
		t.Errorf("expected size %d, got %d", 7, size)
	} else if exp := true; exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// false
	if v, size, err := Bool([]byte{100, 0, 5, 'f', 'a', 'l', 's', 'e'}); err != nil {
		t.Error(err)
	} else if size != 8 {
		t.Errorf("expected size %d, got %d", 8, size)
	} else if exp := false; exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// error
	if _, _, err := Bool([]byte{100, 0, 3, 97, 98, 99}); err == nil {
		t.Error("err == nil")
	} else {
		switch err.(type) {
		case SyntaxError:
		default:
			t.Errorf("error is not SyntaxError, but %T", err)
		}
	}
}

func Test_Float64(t *testing.T) {
	// 0.1
	b := []byte{
		99, 49, 46, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48,
		48, 48, 48, 53, 53, 53, 49, 101, 45, 48, 49, 0, 0, 0, 0, 0,
	}
	if v, size, err := Float64(b); err != nil {
		t.Error(err)
	} else if size != len(b) {
		t.Errorf("expected size %d, got %d", len(b), size)
	} else if exp := 0.1; exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 0.1
	b = []byte{70, 63, 185, 153, 153, 153, 153, 153, 154}
	if v, size, err := Float64(b); err != nil {
		t.Error(err)
	} else if size != len(b) {
		t.Errorf("expected size %d, got %d", len(b), size)
	} else if exp := 0.1; exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// error (31 bytes instead of 32)
	b = []byte{
		99, 49, 46, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48,
		48, 48, 48, 53, 53, 53, 49, 101, 45, 48, 49, 0, 0, 0, 0,
	}
	if _, _, err := Float64(b); err == nil {
		t.Error("err == nil")
	} else if !isStructuralError(err) {
		t.Errorf("error is not StructuralError, but %T", err)
	}

	// error (fail on Sscanf)
	b = []byte{
		99, 99, 46, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48,
		48, 48, 48, 53, 53, 53, 49, 101, 45, 48, 49, 0, 0, 0, 0, 0,
	}
	if _, _, err := Float64(b); err == nil {
		t.Error("err == nil")
	} else if !isStructuralError(err) {
		t.Errorf("error is not StructuralError, but %T", err)
	}

	// error (bad length)
	if _, _, err := Float64([]byte{99}); err == nil {
		t.Error("err == nil")
	} else if !isStructuralError(err) {
		t.Errorf("error is not StructuralError, but %T", err)
	}

	// error (bad length)
	if _, _, err := Float64([]byte{70}); err == nil {
		t.Error("err == nil")
	} else if !isStructuralError(err) {
		t.Errorf("error is not StructuralError, but %T", err)
	}
}

func Test_Int64(t *testing.T) {
	// 255
	if v, size, err := Int64([]byte{97, 255}); err != nil {
		t.Error(err)
	} else if expSize := 2; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := int64(255); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 0x7fffffff
	if v, size, err := Int64([]byte{98, 127, 255, 255, 255}); err != nil {
		t.Error(err)
	} else if expSize := 5; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := int64(0x7fffffff); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// -0x80000000
	if v, size, err := Int64([]byte{98, 128, 0, 0, 0}); err != nil {
		t.Error(err)
	} else if expSize := 5; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := int64(-0x80000000); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// 0x7fffffffffffffff
	if v, size, err := Int64([]byte{110, 8, 0, 255, 255, 255, 255, 255, 255, 255, 127}); err != nil {
		t.Error(err)
	} else if expSize := 11; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := int64(0x7fffffffffffffff); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// -0x8000000000000000
	if v, size, err := Int64([]byte{110, 8, 1, 0, 0, 0, 0, 0, 0, 0, 128}); err != nil {
		t.Error(err)
	} else if expSize := 11; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := int64(-0x8000000000000000); exp != v {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// error (bad length)
	for _, b := range []byte{97, 98, 110, 111} {
		if _, _, err := Int64([]byte{b}); err == nil {
			t.Error("err == nil (%d)", b)
		} else if !isStructuralError(err) {
			t.Errorf("error is not StructuralError, but %T", err)
		}
	}

	// error (0x8000000000000000)
	if _, _, err := Int64([]byte{110, 8, 0, 0, 0, 0, 0, 0, 0, 0, 128}); err == nil {
		t.Error("err == nil")
	} else if !isStructuralError(err) {
		t.Errorf("error is not StructuralError, but %T", err)
	}

	// error (-0x8000000000000001)
	if _, _, err := Int64([]byte{110, 8, 1, 1, 0, 0, 0, 0, 0, 0, 128}); err == nil {
		t.Error("err == nil")
	} else if !isStructuralError(err) {
		t.Errorf("error is not StructuralError, but %T", err)
	}
}

func Test_String(t *testing.T) {
	// "" (nil)
	if v, size, err := String([]byte{106}); err != nil {
		t.Error(err)
	} else if expSize := 1; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := ""; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "" (empty string)
	if v, size, err := String([]byte{107, 0, 0}); err != nil {
		t.Error(err)
	} else if expSize := 3; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := ""; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "" (empty list)
	if v, size, err := String([]byte{108, 0, 0, 0, 0, 106}); err != nil {
		t.Error(err)
	} else if expSize := 6; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := ""; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "" (empty binary)
	if v, size, err := String([]byte{109, 0, 0, 0, 0}); err != nil {
		t.Error(err)
	} else if expSize := 5; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := ""; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "abc"
	if v, size, err := String([]byte{107, 0, 3, 97, 98, 99}); err != nil {
		t.Error(err)
	} else if expSize := 6; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := "abc"; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// <<"abc">>
	if v, size, err := String([]byte{109, 0, 0, 0, 3, 97, 98, 99}); err != nil {
		t.Error(err)
	} else if expSize := 8; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := "abc"; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "фыва", where the last one is of erlSmallBig type
	in := []byte{
		108, 0, 0, 0, 4, 98, 0, 0, 4, 68, 98, 0, 0, 4,
		75, 98, 0, 0, 4, 50, 110, 2, 0, 48, 4, 106,
	}
	if v, size, err := String(in); err != nil {
		t.Error(err)
	} else if expSize := 26; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := "фыва"; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// "фыва", where the last one is of erlLargeBig type
	in = []byte{
		108, 0, 0, 0, 4, 98, 0, 0, 4, 68, 98, 0, 0, 4,
		75, 98, 0, 0, 4, 50, 111, 0, 0, 0, 2, 0, 48, 4, 106,
	}
	if v, size, err := String(in); err != nil {
		t.Error(err)
	} else if expSize := 29; size != expSize {
		t.Errorf("expected size %d, got %d", expSize, size)
	} else if exp := "фыва"; v != exp {
		t.Errorf("expected %v, got %v", exp, v)
	}

	// error (wrong length) in string
	if _, _, err := String([]byte{107, 0, 3, 97, 98}); err == nil {
		t.Error("err == nil")
	} else if !isStructuralError(err) {
		t.Errorf("error is not StructuralError, but %T", err)
	}

	// error (wrong length) in binary string
	if _, _, err := String([]byte{109, 0, 0, 0, 3, 97, 98}); err == nil {
		t.Error("err == nil")
	} else if !isStructuralError(err) {
		t.Errorf("error is not StructuralError, but %T", err)
	}

	// error (improper list) [$a,$b,$c|0]
	if _, _, err := String([]byte{108, 0, 0, 0, 3, 97, 98, 99, 0}); err == nil {
		t.Error("err == nil")
	} else if !isStructuralError(err) {
		t.Errorf("error is not StructuralError, but %T", err)
	}

	// error (bad length)
	for _, b := range []byte{107, 108, 109} {
		if _, _, err := String([]byte{b}); err == nil {
			t.Error("err == nil (%d)", b)
		} else if !isStructuralError(err) {
			t.Errorf("error is not StructuralError, but %T", err)
		}
	}
}
