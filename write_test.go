package etf

import (
	"bytes"
	"github.com/goerlang/etf/parse"
	"github.com/goerlang/etf/types"
	"io"
	"math"
	r "reflect"
	"testing"
)

func testWrite(
	t *testing.T,
	fi, pi, v interface{},
	shouldSize int,
	shouldError bool,
	args ...interface{}) {

	f := func(w io.Writer, data interface{}) interface{} {
		return r.ValueOf(fi).Call([]r.Value{
			r.ValueOf(w),
			r.ValueOf(data),
		})[0].Interface()
	}

	p := func(b []byte) (ret interface{}, size int, err interface{}) {
		result := r.ValueOf(pi).Call([]r.Value{r.ValueOf(b)})
		ret = result[0].Interface()
		size = result[1].Interface().(int)
		err = result[2].Interface()
		return
	}

	var result interface{}
	var resultSize int
	var err interface{}

	w := new(bytes.Buffer)
	w.Reset()
	err = f(w, v)

	if !shouldError {
		// encode
		if err != nil {
			t.Error(err, args)
		} else if shouldSize != w.Len() {
			t.Fatalf("encode %v: expected size %d, got %d. args=%v", args, shouldSize, w.Len())
		}

		// decode
		result, resultSize, err = p(w.Bytes())
		if err != nil {
			t.Error(err, args)
		}
		if shouldSize != resultSize {
			t.Errorf("decode %v: expected size %d, got %d. args=%v", args, shouldSize, resultSize)
		} else {
			switch v.(type) {
			case []byte:
				if bytes.Compare(v.([]byte), result.([]byte)) != 0 {
					t.Errorf("decode %v: expected %v, got %v", v, result)
				}
			default:
				if v != result {
					t.Errorf("decode %v: expected %v, got %v", v, result)
				}
			}
		}
	} else {
		if err == nil {
			t.Error("err == nil", args)
		}

		switch err.(type) {
		case EncodeError:
		default:
			t.Errorf("expected %T, got %T (%#v)", EncodeError{}, err, args)
		}
	}
}

func Test_writeAtom(t *testing.T) {
	testWriteAtom := func(v string, headerSize int, shouldError bool, args ...interface{}) {
		testWrite(t, writeAtom, parse.Atom, types.ErlAtom(v), headerSize+len(v), shouldError, args...)
	}

	testWriteAtom(string(bytes.Repeat([]byte{'a'}, math.MaxUint8+0)), 2, false, "255 $a")
	testWriteAtom(string(bytes.Repeat([]byte{'a'}, math.MaxUint8+1)), 3, false, "256 $a")
	testWriteAtom("", 2, false, "'' (empty atom)")
	testWriteAtom(string(bytes.Repeat([]byte{'a'}, math.MaxUint16+0)), 3, false, "65535 $a")
	testWriteAtom(string(bytes.Repeat([]byte{'a'}, math.MaxUint16+1)), 3, true, "65536 $a")
}

func Test_writeBinary(t *testing.T) {
	testWriteBinary := func(bytes []byte, headerSize int, shouldError bool, args ...interface{}) {
		testWrite(t, writeBinary, parse.Binary, bytes, headerSize+len(bytes), shouldError, args...)
	}

	testWriteBinary([]byte{}, 5, false, "empty binary")
	testWriteBinary(bytes.Repeat([]byte{1}, 64), 5, false, "65535 bytes binary")
}

func Test_writeBool(t *testing.T) {
	testWriteBool := func(b bool, totalSize int, args ...interface{}) {
		testWrite(t, writeBool, parse.Bool, b, totalSize, false, args...)
	}

	testWriteBool(true, 6, "true")
	testWriteBool(false, 7, "false")
}

func Test_writeFloat64(t *testing.T) {
	testWriteFloat64 := func(f float64) {
		testWrite(t, writeFloat64, parse.Float64, f, 9, false, f)
	}

	testWriteFloat64(0.0)
	testWriteFloat64(math.SmallestNonzeroFloat64)
	testWriteFloat64(math.MaxFloat64)
}

func Test_writeInt64_and_BigInt(t *testing.T) {
	testWriteInt64 := func(x int64, totalSize int, shouldError bool, args ...interface{}) {
		testWrite(t, writeInt64, parse.Int64, x, totalSize, shouldError, args...)
	}

	testWriteInt64(0, 2, false, "0")
	testWriteInt64(-1, 5, false, "0")
	testWriteInt64(math.MaxUint8+0, 2, false, "255")
	testWriteInt64(math.MaxUint8+1, 5, false, "256")
	testWriteInt64(math.MaxInt32+0, 5, false, "0x7fffffff")
	testWriteInt64(math.MaxInt32+1, 7, false, "0x80000000")
	testWriteInt64(math.MinInt32+0, 5, false, "-0x80000000")
	testWriteInt64(math.MinInt32-1, 7, false, "-0x80000001")
}

func Test_writeString(t *testing.T) {
	testWriteString := func(v string, headerSize int, shouldError bool, args ...interface{}) {
		testWrite(t, writeString, parse.String, v, headerSize+len(v), shouldError, args...)
	}

	testWriteString(string(bytes.Repeat([]byte{'a'}, math.MaxUint16+0)), 3, false, "65535 $a")
	testWriteString("", 3, false, `"" (empty string)`)
	testWriteString(string(bytes.Repeat([]byte{'a'}, math.MaxUint16+1)), 3, true, "65536 $a")
}
