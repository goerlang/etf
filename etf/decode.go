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
  "fmt"
  . "reflect"
)

var (
  atomType = ValueOf(Atom("")).Type()
)

// OverflowError is returned when number cannot be represented by supplied type.
type OverflowError struct {
  Value interface{}
  Type  Type
}

// TypeError is returned when a type cannot be decoded.
type TypeError struct {
  Type Type
}

// VersionError is returned on invalid Erlang external format version number.
type VersionError struct {
  Version byte
}

func (err OverflowError) Error() string {
  return fmt.Sprintf("overflow error: cannot represent %s by type %s", err.Value, err.Type)
}

func (err TypeError) Error() string {
  return fmt.Sprintf("type error: cannot represent type %s", err.Type)
}

func (err VersionError) Error() string {
  return fmt.Sprintf("version error: version %d is not supported", err.Version)
}

// decodeStruct decodes a structure.
func decodeStruct(b []byte, ptr Value) (size uint, err error) {
  var arity int

  v := ptr.Elem()

  switch erlType(b[0]) {
  case erlSmallTuple:
    // $hA…
    if len(b) >= 2 {
      arity = int(b[1])
      size = 2
      goto decode
    } else {
      err = StructuralError{
        fmt.Sprintf("invalid tuple length (%d)", len(b)),
      }
    }

  case erlLargeTuple:
    // $iAAAA…
    if len(b) >= 5 {
      arity = int(be.Uint32(b[1:5]))
      size = 5
      goto decode
    } else {
      err = StructuralError{
        fmt.Sprintf("invalid tuple length (%d)", len(b)),
      }
    }

  default:
    err = SyntaxError{"not a tuple"}
  }

  return

decode:
  if arity != v.NumField() {
    err = StructuralError{fmt.Sprintf("different number of fields (%d, should be %d)", v.NumField(), arity)}
    return
  }

  for i := 0; i < arity; i++ {
    var s uint
    f := v.Field(i).Addr()

    s, err = decode(b[size:], f)
    size += s

    if err != nil {
      break
    }
  }

  return
}

func decode(b []byte, ptr Value) (size uint, err error) {
  v := ptr.Elem()

  switch v.Kind() {
  case Bool:
    var result bool
    if result, size, err = parseBool(b); err == nil {
      v.SetBool(result)
    }

  case Int, Int8, Int16, Int32, Int64:
    var result int64
    if result, size, err = parseInt64(b); err == nil {
      if v.OverflowInt(result) {
        err = OverflowError{result, v.Type()}
      } else {
        v.SetInt(result)
      }
    }

  case Uint, Uint8, Uint16, Uint32, Uint64:
    var result uint64
    if result, size, err = parseUint64(b); err == nil {
      if v.OverflowUint(result) {
        err = OverflowError{result, v.Type()}
      } else {
        v.SetUint(result)
      }
    }

  case Float32:
    var result float64
    if result, size, err = parseFloat64(b); err == nil {
      if v.OverflowFloat(result) {
        err = OverflowError{result, v.Type()}
      } else {
        v.SetFloat(result)
      }
    }

  case Float64:
    var result float64
    if result, size, err = parseFloat64(b); err == nil {
      v.SetFloat(result)
    }

  case String:
    if v.Type() == atomType {
      var result Atom
      if result, size, err = parseAtom(b); err == nil {
        v.Set(ValueOf(result))
      }
    } else {
      var result string
      if result, size, err = parseString(b); err == nil {
        v.Set(ValueOf(result))
      }
    }

  case Struct:
    size, err = decodeStruct(b, ptr)

  default:
    err = TypeError{v.Type()}
  }

  return
}

// Decode unmarshals a value and stores it to a variable pointer by ptr.
func Decode(b []byte, ptr interface{}) (size uint, err error) {
  if b[0] != erlFormatVersion {
    err = VersionError{b[0]}
  } else {
    p := ValueOf(ptr)
    size, err = decode(b[1:], p)
    size++
  }

  return
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
