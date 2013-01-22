package write

import (
	"bytes"
	"github.com/goerlang/etf/read"
	ty "github.com/goerlang/etf/types"
	"math"
	"testing"
)

func TestAtom(t *testing.T) {
	test := func(in ty.Atom, shouldFail bool) {
		w := new(bytes.Buffer)
		if err := Atom(w, in); err != nil {
			if !shouldFail {
				t.Error(in, err)
			}
		} else if shouldFail {
			t.Error("err == nil (%v)", in)
		} else if v, err := read.Atom(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if v != in {
			t.Errorf("expected %v, got %v", in, v)
		}
	}

	test(ty.Atom(""), false)
	test(ty.Atom(bytes.Repeat([]byte{'a'}, math.MaxUint8)), false)
	test(ty.Atom(bytes.Repeat([]byte{'a'}, math.MaxUint8+1)), false)
	test(ty.Atom(bytes.Repeat([]byte{'a'}, math.MaxUint16)), false)
	test(ty.Atom(bytes.Repeat([]byte{'a'}, math.MaxUint16+1)), true)
}

func TestBinary(t *testing.T) {
	test := func(in []byte) {
		w := new(bytes.Buffer)
		if err := Binary(w, in); err != nil {
			t.Error(in, err)
		} else if v, err := read.Binary(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if bytes.Compare(v, in) != 0 {
			t.Errorf("expected %v, got %v", in, v)
		}
	}

	test([]byte{})
	test(bytes.Repeat([]byte{231}, 65535))
	test(bytes.Repeat([]byte{123}, 65536))
}

func TestBool(t *testing.T) {
	test := func(in bool) {
		w := new(bytes.Buffer)
		if err := Bool(w, in); err != nil {
			t.Error(in, err)
		} else if v, err := read.Bool(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if v != in {
			t.Errorf("expected %v, got %v", in, v)
		}
	}

	test(true)
	test(false)
}

func TestFloat(t *testing.T) {
	test := func(in float64) {
		w := new(bytes.Buffer)
		if err := Float(w, in); err != nil {
			t.Error(in, err)
		} else if v, err := read.Float(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if v != in {
			t.Errorf("expected %v, got %v", in, v)
		}
	}

	test(0.0)
	test(-12345.6789)
	test(math.SmallestNonzeroFloat64)
	test(math.MaxFloat64)
}

func TestInt(t *testing.T) {
	test := func(in int64) {
		w := new(bytes.Buffer)
		if err := Int(w, in); err != nil {
			t.Error(in, err)
		} else if v, err := read.Int(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if v != in {
			t.Errorf("expected %v, got %v", in, v)
		}
	}

	test(0)
	test(-1)
	test(math.MaxInt8)
	test(math.MaxInt8 + 1)
	test(math.MaxInt32)
	test(math.MaxInt32 + 1)
	test(math.MinInt32)
	test(math.MinInt32 - 1)
	test(math.MinInt64)
	test(math.MaxInt64)
}

func TestUint(t *testing.T) {
	test := func(in uint64) {
		w := new(bytes.Buffer)
		if err := Uint(w, in); err != nil {
			t.Error(in, err)
		} else if v, err := read.Uint(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if v != in {
			t.Errorf("expected %v, got %v", in, v)
		}
	}

	test(0)
	test(math.MaxUint8)
	test(math.MaxUint8 + 1)
	test(math.MaxUint32)
	test(math.MaxUint32 + 1)
	test(math.MaxUint64)
}

func TestPid(t *testing.T) {
	test := func(in ty.Pid) {
		w := new(bytes.Buffer)
		if err := Pid(w, in); err != nil {
			t.Error(in, err)
		} else if v, err := read.Pid(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if v != in {
			t.Errorf("expected %v, got %v", in, v)
		}
	}

	test(ty.Pid{ty.Node("omg@lol"), 38, 0, 3})
	test(ty.Pid{ty.Node("self@localhost"), 32, 1, 9})
}

func TestString(t *testing.T) {
	test := func(in string, shouldFail bool) {
		w := new(bytes.Buffer)
		if err := String(w, in); err != nil {
			if !shouldFail {
				t.Error(in, err)
			}
		} else if shouldFail {
			t.Error("err == nil (%v)", in)
		} else if v, err := read.String(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if v != in {
			t.Errorf("expected %v, got %v", in, v)
		}
	}

	test(string(bytes.Repeat([]byte{'a'}, math.MaxUint16)), false)
	test("", false)
	test(string(bytes.Repeat([]byte{'a'}, math.MaxUint16+1)), true)
}

func TestTerm(t *testing.T) {
	type s1 struct {
		L []interface{}
		F float64
	}
	type s2 struct {
		ty.Atom
		S  string
		I  int
		S1 s1
		B  byte
	}
	in := s2{
		ty.Atom("lol"),
		"omg",
		13666,
		s1{
			[]interface{}{
				256,
				"1",
				13.0,
			},
			13.13,
		},
		1,
	}

	w := new(bytes.Buffer)
	if err := Term(w, in); err != nil {
		t.Error(in, err)
	} else {
		if term, err := read.Term(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if err := Term(w, term); err != nil {
			t.Error(term, err)
		} else if term, err := read.Term(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", term, l)
		}
	}
}
