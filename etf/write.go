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
  bin "encoding/binary"
  "fmt"
  "io"
)

// writeBE
func writeBE(w io.Writer, data ...interface{}) error {
  for _, v := range data {
    err := bin.Write(w, be, v)

    if err != nil {
      return err
    }
  }

  return nil
}

// writeAtom
func writeAtom(w io.Writer, a Atom) (err error) {
  size := len(a)

  switch {
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
