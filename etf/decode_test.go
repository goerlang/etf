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
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWA*/

import (
  "github.com/bmizerany/assert"
  "math/big"
  "testing"
)

func Test_parseAtom(t *testing.T) {
  // "abc"
  v, size, err := parseAtom([]byte{100,0,3,97,98,99})
  assert.Equal(t, nil, err)
  assert.Equal(t, Atom("abc"), v)
  assert.Equal(t, uint(6), size)

  // error
  v, size, err = parseAtom([]byte{100,0,4,97,98,99})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatal("error is not StructuralError")
  }
}

func Test_parseBool(t *testing.T) {
  // true
  v, size, err := parseBool([]byte{100,0,4,'t','r','u','e'})
  assert.Equal(t, nil, err)
  assert.T(t, v)
  assert.Equal(t, uint(7), size)

  // false
  v, size, err = parseBool([]byte{100,0,5,'f','a','l','s','e'})
  assert.Equal(t, nil, err)
  assert.T(t, !v)
  assert.Equal(t, uint(8), size)

  // error
  v, size, err = parseBool([]byte{100,0,3,97,98,99})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case SyntaxError:
  default:
    t.Fatal("error is not SyntaxError")
  }
}

func Test_parseInt64(t *testing.T) {
  // 255
  v, size, err := parseInt64([]byte{97,255})
  assert.Equal(t, nil, err)
  assert.Equal(t, int64(255), v)
  assert.Equal(t, uint(2), size)

  // 0x7fffffff
  v, size, err = parseInt64([]byte{98,127,255,255,255})
  assert.Equal(t, nil, err)
  assert.Equal(t, int64(0x7fffffff), v)
  assert.Equal(t, uint(5), size)

  // -0x80000000
  v, size, err = parseInt64([]byte{98,128,0,0,0})
  assert.Equal(t, nil, err)
  assert.Equal(t, int64(-0x80000000), v)
  assert.Equal(t, uint(5), size)

  // 0x7fffffffffffffff
  v, size, err = parseInt64([]byte{110,8,0,255,255,255,255,255,255,255,127})
  assert.Equal(t, nil, err)
  assert.Equal(t, int64(9223372036854775807), v)
  assert.Equal(t, uint(11), size)

  // -0x8000000000000000
  v, size, err = parseInt64([]byte{110,8,1,0,0,0,0,0,0,0,128})
  assert.Equal(t, nil, err)
  assert.Equal(t, int64(-9223372036854775808), v)
  assert.Equal(t, uint(11), size)

  // error (0x8000000000000000)
  v, size, err = parseInt64([]byte{110,8,0,0,0,0,0,0,0,0,128})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatal("error is not StructuralError")
  }

  // error (-0x8000000000000001)
  v, size, err = parseInt64([]byte{110,8,1,1,0,0,0,0,0,0,128})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatal("error is not StructuralError")
  }
}

func Test_parseBigInt(t *testing.T) {
  // (1<<2040)
  b := []byte{
    111,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,1,
  }
  v, size, err := parseBigInt(b)
  assert.Equal(t, nil, err)
  assert.T(t, new(big.Int).Lsh(big.NewInt(1), 2040).Cmp(v) == 0)
  assert.Equal(t, uint(len(b)), size)
}

func Test_parseFloat64(t *testing.T) {
  // 0.1
  v, size, err := parseFloat64([]byte{
    99,49,46,48,48,48,48,48,48,48,48,48,48,48,48,48,
    48,48,48,53,53,53,49,101,45,48,49,0,0,0,0,0,
  })
  assert.Equal(t, nil, err)
  assert.Equal(t, float64(0.1), v)
  assert.Equal(t, uint(32), size)

  // 0.1
  v, size, err = parseFloat64([]byte{70,63,185,153,153,153,153,153,154})
  assert.Equal(t, nil, err)
  assert.Equal(t, float64(0.1), v)
  assert.Equal(t, uint(9), size)

  // error (31 bytes instead of 32)
  v, size, err = parseFloat64([]byte{
    99,49,46,48,48,48,48,48,48,48,48,48,48,48,48,48,
    48,48,48,53,53,53,49,101,45,48,49,0,0,0,0,
  })
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatal("error is not StructuralError")
  }

  // error (fail on Sscanf)
  v, size, err = parseFloat64([]byte{
    99,99,46,48,48,48,48,48,48,48,48,48,48,48,48,48,
    48,48,48,53,53,53,49,101,45,48,49,0,0,0,0,0,
  })
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatal("error is not StructuralError")
  }
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
