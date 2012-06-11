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
  "testing"
)

func Test_writeString(t *testing.T) {
  var size uint
  var v string
  var resultString string
  var resultSize uint
  var err error

  w := new(bytes.Buffer)

  // 65535 'a'
  w.Reset()
  v = string(bytes.Repeat([]byte{'a'}, 65535))
  size = 3 + uint(len(v))
  err = writeString(w, v)
  assert.Equal(t, nil, err)
  assert.Equal(t, size, uint(w.Len()))
  resultString, resultSize, err = parseString(w.Bytes())
  assert.Equal(t, nil, err)
  assert.Equal(t, v, resultString)
  assert.Equal(t, size, resultSize)

  // empty string
  w.Reset()
  v = ""
  size = 3
  err = writeString(w, v)
  assert.Equal(t, nil, err)
  assert.Equal(t, size, uint(w.Len()))
  resultString, resultSize, err = parseString(w.Bytes())
  assert.Equal(t, nil, err)
  assert.Equal(t, v, resultString)
  assert.Equal(t, size, resultSize)

  // error (65536 'a')
  w.Reset()
  v = string(bytes.Repeat([]byte{'a'}, 65536))
  err = writeString(w, v)
  assert.NotEqual(t, nil, err)
  switch err.(type) {
  case EncodeError:
  default:
    t.Fatalf("error is not EncodeError, but %T", err)
  }
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
