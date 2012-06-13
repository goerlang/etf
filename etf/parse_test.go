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
  "math"
  "math/big"
  "math/rand"
  "testing"
  "time"
)

func Benchmark_parseAtom_64(b *testing.B) {
  b.StopTimer()

  rand.Seed(time.Now().UnixNano())
  max := 64
  length := 64
  atoms := make([][]byte, max)

  for i := 0; i < max; i++ {
    w := new(bytes.Buffer)
    s := bytes.Repeat([]byte{'a'}, length)
    writeAtom(w, Atom(string(bytes.Map(func(rune) rune { return rune(byte(rand.Int())) }, s))))
    atoms[i] = w.Bytes()
  }

  b.StartTimer()

  for i := 0; i < b.N; i++ {
    _, _, err := parseAtom(atoms[i % max])

    if err != nil {
      b.Fatal("failed to parse atom")
    }
  }
}

func Test_parseAtom(t *testing.T) {
  // 'abc'
  v, size, err := parseAtom([]byte{100,0,3,97,98,99})
  assert.Equal(t, nil, err)
  assert.Equal(t, Atom("abc"), v)
  assert.Equal(t, uint(6), size)

  // ''
  v, size, err = parseAtom([]byte{100,0,0})
  assert.Equal(t, nil, err)
  assert.Equal(t, Atom(""), v)
  assert.Equal(t, uint(3), size)

  // 'abc' as SmallAtom
  v, size, err = parseAtom([]byte{115,3,97,98,99})
  assert.Equal(t, nil, err)
  assert.Equal(t, Atom("abc"), v)
  assert.Equal(t, uint(5), size)

  // '' as SmallAtom
  v, size, err = parseAtom([]byte{115,0})
  assert.Equal(t, nil, err)
  assert.Equal(t, Atom(""), v)
  assert.Equal(t, uint(2), size)

  // error (ends abruptly)
  v, size, err = parseAtom([]byte{100,0,4,97,98,99})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (bad length)
  v, size, err = parseAtom([]byte{100})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
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
    t.Fatalf("error is not SyntaxError, but %T", err)
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

  // error (bad length)
  v, size, err = parseInt64([]byte{97})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (bad length)
  v, size, err = parseInt64([]byte{98})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (bad length)
  v, size, err = parseInt64([]byte{110})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (bad length)
  v, size, err = parseInt64([]byte{111})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (0x8000000000000000)
  v, size, err = parseInt64([]byte{110,8,0,0,0,0,0,0,0,0,128})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (-0x8000000000000001)
  v, size, err = parseInt64([]byte{110,8,1,1,0,0,0,0,0,0,128})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }
}

func Benchmark_parseBigInt(b *testing.B) {
  b.StopTimer()

  rand := rand.New(rand.NewSource(time.Now().UnixNano()))
  uint64Max := big.NewInt(math.MaxInt64)
  top := new(big.Int).Mul(uint64Max, uint64Max)
  max := 512
  bigints := make([][]byte, max)

  for i := 0; i < max; i++ {
    w := new(bytes.Buffer)
    a := new(big.Int).Rand(rand, top)
    b := new(big.Int).Rand(rand, top)
    writeBigInt(w, new(big.Int).Sub(a, b))
    bigints[i] = w.Bytes()
  }

  b.StartTimer()

  for i := 0; i < b.N; i++ {
    _, _, err := parseBigInt(bigints[i % max])

    if err != nil {
      b.Fatal("failed to parse big int")
    }
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

  // -(1<<2040)
  b = []byte{
    111,0,0,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
    0,0,0,1,
  }
  v, size, err = parseBigInt(b)
  assert.Equal(t, nil, err)
  assert.T(t, new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 2040)).Cmp(v) == 0)
  assert.Equal(t, uint(len(b)), size)
}

func Benchmark_parseFloat64(b *testing.B) {
  b.StopTimer()

  rand.Seed(time.Now().UnixNano())
  max := 512
  floats := make([][]byte, max)

  for i := 0; i < max; i++ {
    w := new(bytes.Buffer)
    writeFloat64(w, rand.ExpFloat64() - rand.ExpFloat64())
    floats[i] = w.Bytes()
  }

  b.StartTimer()

  for i := 0; i < b.N; i++ {
    _, _, err := parseFloat64(floats[i % max])

    if err != nil {
      b.Fatal("failed to parse float")
    }
  }
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
    t.Fatalf("error is not StructuralError, but %T", err)
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
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (bad length)
  v, size, err = parseFloat64([]byte{99})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (bad length)
  v, size, err = parseFloat64([]byte{70})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }
}

func Benchmark_parseString_64(b *testing.B) {
  b.StopTimer()

  rand.Seed(time.Now().UnixNano())
  max := 64
  length := 64
  strings := make([][]byte, max)

  for i := 0; i < max; i++ {
    w := new(bytes.Buffer)
    s := bytes.Repeat([]byte{'a'}, length)
    writeString(w, string(bytes.Map(func(rune) rune { return rune(byte(rand.Int())) }, s)))
    strings[i] = w.Bytes()
  }

  b.StartTimer()

  for i := 0; i < b.N; i++ {
    _, _, err := parseString(strings[i % max])

    if err != nil {
      b.Fatal("failed to parse string")
    }
  }
}

func Test_parseString_and_Binary(t *testing.T) {
  // "" (nil)
  v, size, err := parseString([]byte{106})
  assert.Equal(t, nil, err)
  assert.Equal(t, "", v)
  assert.Equal(t, uint(1), size)

  // "" (empty string)
  v, size, err = parseString([]byte{107,0,0})
  assert.Equal(t, nil, err)
  assert.Equal(t, "", v)
  assert.Equal(t, uint(3), size)

  // "" (empty list)
  v, size, err = parseString([]byte{108,0,0,0,0,106})
  assert.Equal(t, nil, err)
  assert.Equal(t, "", v)
  assert.Equal(t, uint(6), size)

  // "" (empty binary)
  v, size, err = parseString([]byte{109,0,0,0,0})
  assert.Equal(t, nil, err)
  assert.Equal(t, "", v)
  assert.Equal(t, uint(5), size)

  // "abc"
  v, size, err = parseString([]byte{107,0,3,97,98,99})
  assert.Equal(t, nil, err)
  assert.Equal(t, "abc", v)
  assert.Equal(t, uint(6), size)

  // <<"abc">>
  v, size, err = parseString([]byte{109,0,0,0,3,97,98,99})
  assert.Equal(t, nil, err)
  assert.Equal(t, "abc", v)
  assert.Equal(t, uint(8), size)

  // "фыва", where the last one is of erlSmallBig type
  v, size, err = parseString([]byte{
    108,0,0,0,4,98,0,0,4,68,98,0,0,4,
    75,98,0,0,4,50,110,2,0,48,4,106,
  })
  assert.Equal(t, nil, err)
  assert.Equal(t, "фыва", v)
  assert.Equal(t, uint(26), size)

  // "фыва", where the last one is of erlLargeBig type
  v, size, err = parseString([]byte{
    108,0,0,0,4,98,0,0,4,68,98,0,0,4,
    75,98,0,0,4,50,111,0,0,0,2,0,48,4,106,
  })
  assert.Equal(t, nil, err)
  assert.Equal(t, "фыва", v)
  assert.Equal(t, uint(29), size)

  // error (wrong length) in string
  v, size, err = parseString([]byte{107,0,3,97,98})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (wrong length) in binary string
  v, size, err = parseString([]byte{109,0,0,0,3,97,98})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (improper list) [$a,$b,$c|0]
  v, size, err = parseString([]byte{108,0,0,0,3,97,98,99,0})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (bad length)
  v, size, err = parseString([]byte{107})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (bad length)
  v, size, err = parseString([]byte{108})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }

  // error (bad length)
  v, size, err = parseString([]byte{109})
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case StructuralError:
  default:
    t.Fatalf("error is not StructuralError, but %T", err)
  }
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
