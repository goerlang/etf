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
  "encoding/binary"
  "fmt"
  "io"
  "math"
  "math/big"
)

// writeBE
func writeBE(w io.Writer, data ...interface{}) error {
  for _, v := range data {
    err := binary.Write(w, be, v)

    if err != nil {
      return err
    }
  }

  return nil
}

// writeAtom
func writeAtom(w io.Writer, a Atom) (err error) {
  switch size := len(a); {
  case size <= 0xff:
    // $sL…
    err = writeBE(w, be, byte(erlSmallAtom), byte(size), []byte(string(a)))

  case size <= 0xffff:
    // $dLL…
    err = writeBE(w, be, byte(erlAtom), uint16(size), []byte(string(a)))

  default:
    err = EncodeError{fmt.Sprintf("atom is too big (%d bytes)", size)}
  }

  return
}

// writeBigInt
func writeBigInt(w io.Writer, x *big.Int) (err error) {
  sign := 0
  if x.Sign() < 0 {
    sign = 1
  }

  bytes := reverseBytes(x.Abs(x).Bytes())

  switch size := len(bytes); {
  case size <= 0xff:
    // $nAS…
    err = writeBE(w, be, byte(erlSmallBig), byte(size), byte(sign), bytes)

  case int(uint32(size)) == size:
    // $oAAAAS…
    err = writeBE(w, be, byte(erlLargeBig), uint32(size), byte(sign), bytes)

  default:
    err = EncodeError{fmt.Sprintf("bad big int size (%d)", size)}
  }

  return
}

// writeBool
func writeBool(w io.Writer, b bool) (err error) {
  switch b {
  case true:
    err = writeAtom(w, Atom("true"))

  case false:
    err = writeAtom(w, Atom("false"))
  }

  return
}

// writeFloat64
func writeFloat64(w io.Writer, f float64) error {
  return writeBE(w, be, byte(erlNewFloat), math.Float64bits(f))
}

// writeInt64
func writeInt64(w io.Writer, x int64) (err error) {
  switch {
  case x >= 0 && x <= 0xff:
    // $aI
    err = writeBE(w, be, byte(erlSmallInteger), byte(x))

  case int64(int32(x)) == x:
    // $bIIII
    err = writeBE(w, be, byte(erlInteger), int32(x))

  default:
    err = writeBigInt(w, big.NewInt(x))
  }

  return
}

// writeString
func writeString(w io.Writer, s string) (err error) {
  switch size := len(s); {
  case size <= 0xffff:
    // $kLL…
    err = writeBE(w, byte(erlString), uint16(size), []byte(s))

  default:
    err = EncodeError{fmt.Sprintf("string is too big (%d bytes)", size)}
  }

  return
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
