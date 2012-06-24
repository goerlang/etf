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
  "github.com/bmizerany/assert"
  "math/big"
  "testing"
)

func Test_Decode_BigInt(t *testing.T) {
  var bigint *big.Int

  size, err := Decode([]byte{131,110,15,0,0,0,0,0,16,
    159,75,179,21,7,201,123,206,151,192,
  }, &bigint)
  assert.Equal(t, nil, err)
  assert.Equal(t, uint(19), size)
}

func Test_Decode(t *testing.T) {
  var s string
  size, err := Decode([]byte{131,107,0,3,49,50,51}, &s)
  assert.Equal(t, nil, err)
  assert.Equal(t, uint(7), size)
  assert.Equal(t, "123", s)

  type testStruct struct {
    Atom
    X uint8
    S string
  }

  var ts testStruct

  size, err = Decode([]byte{
    131,104,3,100,0,4,98,108,97,104,97,4,108,0,0,0,4,98,
    0,0,4,68,98,0,0,4,75,98,0,0,4,50,98,0,0,4,48,106,
  }, &ts)
  assert.Equal(t, nil, err)
  assert.Equal(t, uint(38), size)
  assert.Equal(t, uint8(4), ts.X)
  assert.Equal(t, "фыва", ts.S)

  size, err = Decode([]byte{
    131,104,3,99,50,46,57,57,57,57,57,57,57,57,57,57,57,57,57,
    57,57,56,56,56,57,56,101,45,48,49,0,0,0,0,0,97,4,108,0,0,
    0,4,98,0,0,4,68,98,0,0,4,75,98,0,0,4,50,98,0,0,4,48,106,
  }, &ts)
  assert.NotEqual(t, nil, err)

  type testStruct2 struct {
    T testStruct
    Y int
  }

  var ts2 testStruct2

  size, err = Decode([]byte{
    131,104,2,104,3,100,0,4,98,108,97,104,97,4,108,0,0,0,4,98,
    0,0,4,68,98,0,0,4,75,98,0,0,4,50,98,0,0,4,48,106,98,0,0,2,154,
  }, &ts2)
  assert.Equal(t, nil, err)
  assert.Equal(t, uint(45), size)
  assert.Equal(t, uint8(4), ts2.T.X)
  assert.Equal(t, "фыва", ts2.T.S)
  assert.Equal(t, 666, ts2.Y)
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
