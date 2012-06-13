package etf

/*
Copyright © 2012 Serge Zirukin

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

import (
  "bytes"
  "github.com/bmizerany/assert"
  "io"
  "math"
  r "reflect"
  "testing"
)

func testWrite(
  t *testing.T,
  fi, pi, v interface{},
  shouldSize uint,
  shouldError bool,
  args ...interface{}) {

  f := func(w io.Writer, data interface{}) interface{} {
    return r.ValueOf(fi).Call([]r.Value{
      r.ValueOf(w),
      r.ValueOf(data),
    })[0].Interface()
  }

  p := func(b []byte) (ret interface{}, size uint, err interface{}) {
    result := r.ValueOf(pi).Call([]r.Value{r.ValueOf(b)})
    ret = result[0].Interface()
    size = result[1].Interface().(uint)
    err = result[2].Interface()
    return
  }

  var result interface{}
  var resultSize uint
  var err interface{}

  w := new(bytes.Buffer)
  w.Reset()
  err = f(w, v)

  if !shouldError {
    assert.Equal(t, nil, err, args...)
    assert.Equalf(t, shouldSize, uint(w.Len()), "encode %v", args)
    result, resultSize, err = p(w.Bytes())
    assert.Equal(t, nil, err, args...)
    assert.Equal(t, v, result, args...)
    assert.Equalf(t, shouldSize, resultSize, "decode %v", args)
  } else {
    assert.NotEqual(t, nil, err, args...)
    switch err.(type) {
    case EncodeError:
    default:
      t.Fatalf("error is not EncodeError, but %T (%#v)", err, args)
    }
  }
}

func Test_writeAtom(t *testing.T) {
  testWriteAtom := func(v string, headerSize uint, shouldError bool, args ...interface{}) {
    testWrite(t, writeAtom, parseAtom, Atom(v), headerSize + uint(len(v)), shouldError, args...)
  }

  testWriteAtom(string(bytes.Repeat([]byte{'a'}, math.MaxUint8 + 0)), 2, false, "255 $a")
  testWriteAtom(string(bytes.Repeat([]byte{'a'}, math.MaxUint8 + 1)), 3, false, "256 $a")
  testWriteAtom("", 2, false, "'' (empty atom)")
  testWriteAtom(string(bytes.Repeat([]byte{'a'}, math.MaxUint16 + 0)), 3, false, "65535 $a")
  testWriteAtom(string(bytes.Repeat([]byte{'a'}, math.MaxUint16 + 1)), 3, true, "65536 $a")
}

func Test_writeBinary(t *testing.T) {
  testWriteBinary := func(bytes []byte, headerSize uint, shouldError bool, args ...interface{}) {
    testWrite(t, writeBinary, parseBinary, bytes, headerSize + uint(len(bytes)), shouldError, args...)
  }

  testWriteBinary([]byte{}, 5, false, "empty binary")
  testWriteBinary(bytes.Repeat([]byte{1}, 64), 5, false, "65535 bytes binary")
}

func Test_writeBool(t *testing.T) {
  testWriteBool := func(b bool, totalSize uint, args ...interface{}) {
    testWrite(t, writeBool, parseBool, b, totalSize, false, args...)
  }

  testWriteBool(true, 6, "true")
  testWriteBool(false, 7, "false")
}

func Test_writeFloat64(t *testing.T) {
  testWriteFloat64 := func(f float64) {
    testWrite(t, writeFloat64, parseFloat64, f, 9, false, f)
  }

  testWriteFloat64(0.0)
  testWriteFloat64(math.SmallestNonzeroFloat64)
  testWriteFloat64(math.MaxFloat64)
}

func Test_writeInt64_and_BigInt(t *testing.T) {
  testWriteInt64 := func(x int64, totalSize uint, shouldError bool, args ...interface{}) {
    testWrite(t, writeInt64, parseInt64, x, totalSize, shouldError, args...)
  }

  testWriteInt64(0, 2, false, "0")
  testWriteInt64(-1, 5, false, "0")
  testWriteInt64(math.MaxUint8 + 0, 2, false, "255")
  testWriteInt64(math.MaxUint8 + 1, 5, false, "256")
  testWriteInt64(math.MaxInt32 + 0, 5, false, "0x7fffffff")
  testWriteInt64(math.MaxInt32 + 1, 7, false, "0x80000000")
  testWriteInt64(math.MinInt32 + 0, 5, false, "-0x80000000")
  testWriteInt64(math.MinInt32 - 1, 7, false, "-0x80000001")
}

func Test_writeString(t *testing.T) {
  testWriteString := func(v string, headerSize uint, shouldError bool, args ...interface{}) {
    testWrite(t, writeString, parseString, v, headerSize + uint(len(v)), shouldError, args...)
  }

  testWriteString(string(bytes.Repeat([]byte{'a'}, math.MaxUint16 + 0)), 3, false, "65535 $a")
  testWriteString("", 3, false, `"" (empty string)`)
  testWriteString(string(bytes.Repeat([]byte{'a'}, math.MaxUint16 + 1)), 3, true, "65536 $a")
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
