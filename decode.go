package etf

import (
	"fmt"
	"math/big"
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
	return fmt.Sprintf(
		"overflow error: cannot represent %s by type %s",
		err.Value,
		err.Type,
	)
}

func (err TypeError) Error() string {
	return fmt.Sprintf("type error: cannot represent type %s", err.Type)
}

func (err VersionError) Error() string {
	return fmt.Sprintf("version error: version %d is not supported", err.Version)
}

// decodeStruct decodes a structure from tuple.
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
		err = StructuralError{
			fmt.Sprintf(
				"different number of fields (%d, should be %d)",
				v.NumField(),
				arity,
			),
		}
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

// decodePtr decodes a specific type of data.
func decodePtr(b []byte, ptr Value) (size uint, err error) {
	v := ptr.Elem()

	switch v.Interface().(type) {
	case *big.Int:
		var result *big.Int
		if result, size, err = parseBigInt(b); err == nil {
			v.Set(ValueOf(result))
		}

	default:
		err = TypeError{v.Type()}
	}

	return
}

// decodeSlice decodes a slice.
func decodeSlice(b []byte, ptr Value) (size uint, err error) {
	v := ptr.Elem()

	switch v.Interface().(type) {
	case []byte:
		var result []byte
		if result, size, err = parseBinary(b); err == nil {
			v.SetBytes(result)
		}

	default:
		err = TypeError{v.Type()}
	}

	return
}

// decodeArray decodes an array.
func decodeArray(b []byte, ptr Value) (size uint, err error) {
	v := ptr.Elem()
	length := v.Len()

	switch v.Type().Elem().Kind() {
	case Uint8:
		var result []byte
		if result, size, err = parseBinary(b); err == nil {
			if length == len(result) {
				for i := range result {
					v.Index(i).SetUint(uint64(result[i]))
				}
			} else {
				err = OverflowError{result, v.Type()}
			}
		}

	default:
		err = TypeError{v.Type()}
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

	case Ptr:
		size, err = decodePtr(b, ptr)

	case Slice:
		size, err = decodeSlice(b, ptr)

	case Array:
		size, err = decodeArray(b, ptr)

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
