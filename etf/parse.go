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
  "math"
  "math/big"
)

var be = bin.BigEndian

const (
  errPrefix = "Erlang external format "
)

type StructuralError struct {
  Msg string
}

type SyntaxError struct {
  Msg string
}

func (err StructuralError) Error() string {
  return errPrefix + "structural error: " + err.Msg
}

func (err SyntaxError) Error() string {
  return errPrefix + "syntax error: " + err.Msg
}

// parseAtom
func parseAtom(b []byte) (ret Atom, size uint, err error) {
  switch erlType(b[0]) {
  case erlAtom:
    // $dLL…
    if len(b) > 3 {
      size = 3 + uint(be.Uint16(b[1:3]))

      if uint(len(b)) >= size {
        ret = Atom(b[3:size])
      } else {
        err = StructuralError{"wrong atom size"}
      }
    }

  case erlSmallAtom:
    // $sL…
    if len(b) > 2 {
      size = 2 + uint(b[1])

      if uint(len(b)) >= size {
        ret = Atom(b[2:size])
      } else {
        err = StructuralError{"wrong atom size"}
      }
    }

  default:
    err = SyntaxError{"not an atom"}
  }

  return
}

// parseBigInt
func parseBigInt(b []byte) (ret *big.Int, size uint, err error) {
  switch erlType(b[0]) {
  case erlSmallBig:
    // $nAS…
    if len(b) > 3 && len(b) - 3 >= int(b[1]) {
      size = 3 + uint(b[1])
      b2 := reverseBytes(b[3:int(3 + b[1])])
      ret = new(big.Int).SetBytes(b2)

      if b[2] != 0 {
        ret = ret.Neg(ret)
      }

      return
    }

  case erlLargeBig:
    // $oAAAAS…
    if len(b) > 6 {
      length := uint(be.Uint32(b[1:5]))
      size = 6 + length

      if uint(len(b)) >= size && uint(int(length) + 6) == length + 6 {
        b2 := reverseBytes(b[6:6 + int(length)])
        ret = new(big.Int).SetBytes(b2)

        if b[5] != 0 {
          ret = ret.Neg(ret)
        }

        return
      } else {
        err = StructuralError{"invalid large bigInt"}
      }
    }
  }

  err = SyntaxError{"not a big integer"}

  return
}

// parseBool
func parseBool(b []byte) (ret bool, size uint, err error) {
  var v Atom

  v, size, err = parseAtom(b)

  if err == nil {
    switch v {
    case Atom("true"):
      ret = true
      return

    case Atom("false"):
      ret = false
      return
    }

    err = SyntaxError{"not a boolean"}
  }

  return
}

// parseFloat64
func parseFloat64(b []byte) (ret float64, size uint, err error) {
  switch erlType(b[0]) {
  case erlFloat:
    // $cFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF0
    if len(b) >= 32 {
      var r int
      r, err = fmt.Sscanf(string(b[1:32]), "%f", &ret)
      size += 32

      if r != 1 || err != nil {
        err = StructuralError{"failed to scan float"}
      }
    } else {
      err = StructuralError{fmt.Sprintf("invalid float length (%d)", len(b))}
    }

  case erlNewFloat:
    // $FFFFFFFFF
    if len(b) >= 9 {
      ret = math.Float64frombits(be.Uint64(b[1:9]))
      size += 9
    } else {
      err = StructuralError{fmt.Sprintf("invalid float length (%d)", len(b))}
    }

  default:
    err = SyntaxError{"not a float"}
  }

  return
}

// parseInt64
func parseInt64(b []byte) (ret int64, size uint, err error) {
  switch erlType(b[0]) {
  case erlSmallInteger:
    // $aI
    if len(b) >= 2 {
      return int64(b[1]), 2, nil
    }

  case erlInteger:
    // $bIIII
    if len(b) >= 5 {
      return int64(int(be.Uint32(b[1:5]))), 5, nil
    }

  case erlSmallBig, erlLargeBig:
    var v *big.Int
    v, size, err = parseBigInt(b)

    if err == nil {
      ret = v.Int64()

      if v.Cmp(big.NewInt(ret)) != 0 {
        err = StructuralError{"integer too large"}
      }
    }

    return
  }

  err = SyntaxError{"not an integer"}

  return
}

// parseString
func parseString(b []byte) (ret string, size uint, err error) {
  switch erlType(b[0]) {
  case erlString:
    // $kLL…
    if len(b) > 3 {
      size = 3 + uint(be.Uint16(b[1:3]))

      if uint(len(b)) >= size {
        ret = string(b[3:size])
      } else {
        err = StructuralError{
          fmt.Sprintf("invalid string size (%d), len=%d", size, len(b)),
        }
      }
    } else {
      err = StructuralError{
        fmt.Sprintf("invalid string length (%d)", len(b)),
      }
    }

  case erlBinary:
    // $mLLLL…
    if len(b) >= 5 {
      size = 5 + uint(be.Uint32(b[1:5]))

      if uint(len(b)) >= size {
        ret = string(b[5:size])
      } else {
        err = StructuralError{
          fmt.Sprintf("invalid binary size (%d), len=%d", size, len(b)),
        }
      }
    } else {
      err = StructuralError{
        fmt.Sprintf("invalid binary length (%d)", len(b)),
      }
    }

  case erlList:
    // $lLLLL…$j
    if len(b) >= 5 {
      strLen := uint(be.Uint32(b[1:5]))
      size = 5
      b = b[size:]

      err = StructuralError{"not a string"}

      for i := uint(0); i < strLen; i++ {
        if len(b) <= 0 {
          err = StructuralError{"string ends abruptly"}
          return
        }

        switch erlType(b[0]) {
        case erlSmallInteger, erlInteger, erlSmallBig, erlLargeBig:
          var char int64
          var charSize uint
          var charErr error
          char, charSize, charErr = parseInt64(b)

          if charErr == nil {
            ret += string(char)
            size += charSize
            b = b[charSize:]
          } else {
            return
          }

        default:
          return
        }
      }

      if len(b) < 1 || erlType(b[0]) != erlNil {
        err = StructuralError{"not a string (got improper list)"}
        return
      }

      size++
      err = nil
    }

  case erlNil:
    // $j
    size++

  default:
    err = SyntaxError{"not a string"}
  }

  return
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
