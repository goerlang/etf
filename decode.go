package etf

import (
	"encoding/binary"
	"errors"
	"fmt"
	p "github.com/goerlang/etf/parse"
	t "github.com/goerlang/etf/types"
	"math/big"
	r "reflect"
)

var (
	atomType     = r.ValueOf(t.ErlAtom("")).Type()
	ErrBadFormat = errors.New("etf: bad format")
)

// OverflowError is returned when number cannot be represented by supplied type.
type OverflowError struct {
	Value interface{}
	Type  r.Type
}

// TypeError is returned when a type cannot be decoded.
type TypeError struct {
	Type r.Type
}

// VersionError is returned on invalid Erlang external format version number.
type VersionError struct {
	Version byte
}

// Decode unmarshals a value and stores it to a variable pointer by ptr.
func Decode(b []byte, ptr interface{}) (size int, err error) {
	if len(b) < 1 {
		return 0, ErrBadFormat
	} else if b[0] != t.ErlFormatVersion {
		err = VersionError{b[0]}
	} else {
		p := r.ValueOf(ptr)
		size, err = decode(b[1:], p)
		size++
	}

	return
}

// DecodeOneOf tries to unmarshal a value to one of the variables.
// It also returns the value which was set successfully.
func DecodeOneOf(b []byte, ptrs ...interface{}) (v interface{}, size int, err error) {
	for _, ptr := range ptrs {
		if size, err = Decode(b, ptr); err == nil {
			v = ptr
			break
		}
	}

	return
}

func decode(b []byte, ptr r.Value) (size int, err error) {
	v := ptr.Elem()

	switch v.Kind() {
	case r.Bool:
		var result bool
		if result, size, err = p.Bool(b); err == nil {
			v.SetBool(result)
		}
	case r.Int, r.Int8, r.Int16, r.Int32, r.Int64:
		var result int64
		if result, size, err = p.Int64(b); err != nil {
			break
		}
		if v.OverflowInt(result) {
			err = OverflowError{result, v.Type()}
		} else {
			v.SetInt(result)
		}
	case r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64, r.Uintptr:
		var result uint64
		if result, size, err = p.UInt64(b); err != nil {
			break
		}
		if v.OverflowUint(result) {
			err = OverflowError{result, v.Type()}
		} else {
			v.SetUint(result)
		}
	case r.Float32, r.Float64:
		var result float64
		if result, size, err = p.Float64(b); err != nil {
			break
		}
		if v.OverflowFloat(result) {
			err = OverflowError{result, v.Type()}
		} else {
			v.SetFloat(result)
		}
	case r.Interface:
	case r.Map:
	case r.Ptr:
		size, err = decodeSpecial(b, v)
	case r.String:
		if v.Type() == atomType {
			var result t.ErlAtom
			if result, size, err = p.Atom(b); err == nil {
				v.Set(r.ValueOf(result))
			}
		} else {
			var result string
			if result, size, err = p.String(b); err == nil {
				v.Set(r.ValueOf(result))
			}
		}
	case r.Array:
		size, err = decodeArray(b, v)
	case r.Slice:
		size, err = decodeSlice(b, v)
	case r.Struct:
		size, err = decodeStruct(b, v)
	default:
		err = TypeError{v.Type()}
	}

	return
}

func decodeArray(b []byte, v r.Value) (size int, err error) {
	length := v.Len()

	switch v.Type().Elem().Kind() {
	case r.Uint8:
		var result []byte
		if result, size, err = p.Binary(b); err == nil {
			if length == len(result) {
				for i := range result {
					v.Index(i).SetUint(uint64(result[i]))
				}
			} else {
				err = OverflowError{result, v.Type()}
			}
		}

	default:
		size, err = decodeList(b, v)
	}

	return
}

func decodeList(b []byte, v r.Value) (size int, err error) {
	switch v.Kind() {
	case r.Slice, r.Array:
		switch b[0] {
		case t.ErlTypeList:
			// $lLLLL…$j
			if len(b) <= 5 {
				err = p.StructuralError{
					fmt.Sprintf("invalid list length (%d)", len(b)),
				}
				break
			}

			listLen := uint(binary.BigEndian.Uint32(b[1:5]))
			size = 5
			b = b[size:]

			slice := r.MakeSlice(v.Type(), int(listLen), int(listLen))
			for i := uint(0); i < listLen; i++ {
				if elemSize, err := decode(b, slice.Index(int(i)).Addr()); err == nil {
					size += elemSize
					b = b[elemSize:]
				} else {
					break
				}
			}

			if len(b) < 1 || t.ErlType(b[0]) != t.ErlTypeNil {
				err = p.StructuralError{"got improper list"}
			} else {
				size++
				v.Set(slice)
			}

		case t.ErlTypeNil:
			// empty slice -- do not touch it
			return 1, nil
		}

	default:
		err = TypeError{v.Type()}
	}

	return
}

func decodeSlice(b []byte, v r.Value) (size int, err error) {
	switch v.Interface().(type) {
	case []byte:
		var result []byte
		if result, size, err = p.Binary(b); err == nil {
			v.SetBytes(result)
		}

	default:
		size, err = decodeList(b, v)
	}

	return
}

func decodeSpecial(b []byte, v r.Value) (size int, err error) {
	switch v.Interface().(type) {
	case *big.Int:
		var result *big.Int
		if result, size, err = p.BigInt(b); err == nil {
			v.Set(r.ValueOf(result))
		}

	default:
		err = TypeError{v.Type()}
	}

	return
}

func decodeStruct(b []byte, v r.Value) (size int, err error) {
	var arity int

	switch t.ErlType(b[0]) {
	case t.ErlTypeSmallTuple:
		// $hA…
		if len(b) >= 2 {
			arity = int(b[1])
			size = 2
			goto decode
		} else {
			err = p.StructuralError{
				fmt.Sprintf("invalid tuple length (%d)", len(b)),
			}
		}

	case t.ErlTypeLargeTuple:
		// $iAAAA…
		if len(b) >= 5 {
			arity = int(binary.BigEndian.Uint32(b[1:5]))
			size = 5
			goto decode
		} else {
			err = p.StructuralError{
				fmt.Sprintf("invalid tuple length (%d)", len(b)),
			}
		}

	default:
		err = p.SyntaxError{"not a tuple"}
	}

	return

decode:
	var fieldsSet int
	for i := 0; i < v.NumField(); i++ {
		var s int
		f := v.Field(i)
		if f.CanSet() {
			s, err = decode(b[size:], f.Addr())
			size += s
			fieldsSet++

			if err != nil {
				break
			}
		}
	}

	if arity != fieldsSet {
		err = p.StructuralError{
			fmt.Sprintf(
				"different number of fields (%d, should be %d)",
				v.NumField(),
				arity,
			),
		}
		return
	}

	return
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
