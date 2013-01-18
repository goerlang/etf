package write

import (
	"bytes"
	"github.com/goerlang/etf/parse"
	. "github.com/goerlang/etf/types"
	"math"
	"testing"
)

func TestAtom(t *testing.T) {
	test := func(in ErlAtom, shouldFail bool) {
		w := new(bytes.Buffer)
		if err := Atom(w, in); err != nil {
			if !shouldFail {
				t.Error(in, err)
			}
		} else if shouldFail {
			t.Error("err == nil (%v)", in)
		} else if v, err := parse.Atom(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if v != in {
			t.Errorf("expected %v, got %v", in, v)
		}
	}

	test(ErlAtom(""), false)
	test(ErlAtom(bytes.Repeat([]byte{'a'}, math.MaxUint8)), false)
	test(ErlAtom(bytes.Repeat([]byte{'a'}, math.MaxUint8+1)), false)
	test(ErlAtom(bytes.Repeat([]byte{'a'}, math.MaxUint16)), false)
	test(ErlAtom(bytes.Repeat([]byte{'a'}, math.MaxUint16+1)), true)
}

func TestBinary(t *testing.T) {
	test := func(in []byte) {
		w := new(bytes.Buffer)
		if err := Binary(w, in); err != nil {
			t.Error(in, err)
		} else if v, err := parse.Binary(w); err != nil {
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
		} else if v, err := parse.Bool(w); err != nil {
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

func TestFloat64(t *testing.T) {
	test := func(in float64) {
		w := new(bytes.Buffer)
		if err := Float64(w, in); err != nil {
			t.Error(in, err)
		} else if v, err := parse.Float64(w); err != nil {
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

func TestInt64(t *testing.T) {
	test := func(in int64) {
		w := new(bytes.Buffer)
		if err := Int64(w, in); err != nil {
			t.Error(in, err)
		} else if v, err := parse.Int64(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if v != in {
			t.Errorf("expected %v, got %v", in, v)
		}
	}

	test(0)
	test(-1)
	test(math.MaxUint8 + 0)
	test(math.MaxUint8 + 1)
	test(math.MaxInt32 + 0)
	test(math.MaxInt32 + 1)
	test(math.MinInt32 + 0)
	test(math.MinInt32 - 1)
}

func TestPid(t *testing.T) {
	test := func(in ErlPid) {
		w := new(bytes.Buffer)
		if err := Pid(w, in); err != nil {
			t.Error(in, err)
		} else if v, err := parse.Pid(w); err != nil {
			t.Error(in, err)
		} else if l := w.Len(); l != 0 {
			t.Errorf("%v: buffer len %d", in, l)
		} else if v != in {
			t.Errorf("expected %v, got %v", in, v)
		}
	}

	test(ErlPid{Node("omg@lol"), 38, 0, 3})
	test(ErlPid{Node("self@localhost"), 32, 1, 9})
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
		} else if v, err := parse.String(w); err != nil {
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
