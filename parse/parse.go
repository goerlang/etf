package parse

import (
	"encoding/binary"
	"errors"
	"fmt"
	. "github.com/goerlang/etf/types"
	"io"
	"math"
	"math/big"
)

type ErrTypeDiffer struct {
	Got byte
	Exp []byte
}

var be = binary.BigEndian

var (
	ErrFloatScan    = errors.New("parse: failed to sscanf float")
	ErrImproperList = errors.New("parse: improper list")
	ErrIntTooBig    = errors.New("parse: integer too big")
	ErrBadBoolean   = errors.New("parse: invalid boolean")
)

func (e *ErrTypeDiffer) Error() string {
	exp := make([]string, len(e.Exp))
	for i, v := range e.Exp {
		exp[i] = fmt.Sprintf("%s(%d)", TypeName(v), v)
	}
	return fmt.Sprintf("parse: type expected one of %s, got %s(%d)",
		exp, TypeName(e.Got), e.Got,
	)
}

func Atom(r io.Reader) (ret ErlAtom, err error) {
	etype, err := termType(r)
	if err != nil {
		return
	}

	switch etype {
	case ErlTypeAtom:
		// $dLL…
		var size uint16
		if err = binary.Read(r, binary.BigEndian, &size); err == nil {
			b := make([]byte, int(size))
			_, err = io.ReadFull(r, b)
			ret = ErlAtom(b)
		}

	case ErlTypeSmallAtom:
		// $sL…
		var size uint8
		if err = binary.Read(r, binary.BigEndian, &size); err == nil {
			b := make([]byte, int(size))
			_, err = io.ReadFull(r, b)
			ret = ErlAtom(b)
		}

	default:
		err = &ErrTypeDiffer{etype, []byte{ErlTypeAtom, ErlTypeSmallAtom}}
	}

	return
}

func BigInt(r io.Reader) (ret *big.Int, err error) {
	etype, err := termType(r)
	if err == nil {
		ret, err = getBigInt(etype, r)
	}

	return
}

func Binary(r io.Reader) (ret []byte, err error) {
	etype, err := termType(r)
	if err == nil {
		ret, err = getBinary(etype, r)
	}

	return
}

func Bool(r io.Reader) (ret bool, err error) {
	v, err := Atom(r)
	if err != nil {
		return
	}

	switch v {
	case ErlAtom("true"):
		ret = true

	case ErlAtom("false"):
		ret = false

	default:
		err = ErrBadBoolean
	}

	return
}

func Float64(r io.Reader) (ret float64, err error) {
	etype, err := termType(r)
	if err != nil {
		return
	}

	switch etype {
	case ErlTypeFloat:
		// $cFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF0
		b := make([]byte, 31)
		if _, err = io.ReadFull(r, b); err == nil {
			var r int
			if r, err = fmt.Sscanf(string(b), "%f", &ret); r != 1 && err == nil {
				err = ErrFloatScan
			}
		}

	case ErlTypeNewFloat:
		// $FFFFFFFFF
		b := make([]byte, 8)
		if _, err = io.ReadFull(r, b); err == nil {
			ret = math.Float64frombits(be.Uint64(b))
		}

	default:
		err = &ErrTypeDiffer{etype, []byte{ErlTypeFloat, ErlTypeNewFloat}}
	}

	return
}

func Int64(r io.Reader) (ret int64, err error) {
	etype, err := termType(r)
	if err == nil {
		ret, err = getInt64(etype, r)
	}

	return
}

func UInt64(r io.Reader) (ret uint64, err error) {
	iret, err := Int64(r)
	ret = uint64(iret)
	return
}

func String(r io.Reader) (ret string, err error) {
	etype, err := termType(r)
	if err != nil {
		return
	}

	switch etype {
	case ErlTypeString, ErlTypeBinary:
		var b []byte
		b, err = getBinary(etype, r)
		ret = string(b)

	case ErlTypeList:
		// $lLLLL…$j
		var size uint32
		if err = binary.Read(r, binary.BigEndian, &size); err != nil {
			return
		}

		b := make([]byte, 1)

		for i := uint32(0); i < size; i++ {
			if _, err = io.ReadFull(r, b); err != nil {
				return
			}

			etype = b[0]
			switch etype {
			case ErlTypeSmallInteger, ErlTypeInteger, ErlTypeSmallBig, ErlTypeLargeBig:
				var char int64
				if char, err = getInt64(etype, r); err != nil {
					return
				}

				ret += string(char)

			default:
				err = &ErrTypeDiffer{
					etype,
					[]byte{
						ErlTypeSmallInteger,
						ErlTypeInteger,
						ErlTypeSmallBig,
						ErlTypeLargeBig,
					},
				}
				return
			}
		}

		if _, err = io.ReadFull(r, b); err == nil && b[0] != ErlTypeNil {
			err = ErrImproperList
		}

	case ErlTypeNil:
		// $j

	default:
		err = &ErrTypeDiffer{
			etype,
			[]byte{
				ErlTypeString,
				ErlTypeBinary,
				ErlTypeList,
				ErlTypeNil,
			},
		}
	}

	return
}

func getBigInt(etype byte, r io.Reader) (ret *big.Int, err error) {
	var size uint32
	var sign byte

	switch etype {
	case ErlTypeSmallBig:
		// $nAS…
		b := make([]byte, 2)
		if _, err = io.ReadFull(r, b); err == nil {
			size = uint32(b[0])
			sign = b[1]
		}

	case ErlTypeLargeBig:
		// $oAAAAS…
		b := make([]byte, 5)
		if _, err = io.ReadFull(r, b); err == nil {
			size = binary.BigEndian.Uint32(b[:4])
			sign = b[4]
		}

	default:
		err = &ErrTypeDiffer{etype, []byte{ErlTypeSmallBig, ErlTypeLargeBig}}
	}

	if err == nil {
		b := make([]byte, int(size))
		if _, err = io.ReadFull(r, b); err == nil {
			ret = new(big.Int).SetBytes(reverse(b))

			if sign != 0 {
				ret = ret.Neg(ret)
			}
		}
	}

	return
}

func getBinary(etype byte, r io.Reader) (ret []byte, err error) {
	switch etype {
	case ErlTypeBinary:
		// $mLLLL…
		var size uint32
		if err = binary.Read(r, binary.BigEndian, &size); err == nil {
			ret = make([]byte, size)
			_, err = io.ReadFull(r, ret)
		}

	case ErlTypeString:
		// $kLL…
		var size uint16
		if err = binary.Read(r, binary.BigEndian, &size); err == nil {
			ret = make([]byte, size)
			_, err = io.ReadFull(r, ret)
		}

	default:
		err = &ErrTypeDiffer{etype, []byte{ErlTypeBinary, ErlTypeString}}
	}

	return
}

func getInt64(etype byte, r io.Reader) (ret int64, err error) {
	switch etype {
	case ErlTypeSmallInteger:
		// $aI
		var x uint8
		err = binary.Read(r, binary.BigEndian, &x)
		ret = int64(x)

	case ErlTypeInteger:
		// $bIIII
		var x int32
		err = binary.Read(r, binary.BigEndian, &x)
		ret = int64(x)

	case ErlTypeSmallBig, ErlTypeLargeBig:
		var v *big.Int
		if v, err = getBigInt(etype, r); err == nil {
			ret = v.Int64()

			if v.Cmp(big.NewInt(ret)) != 0 {
				err = ErrIntTooBig
			}
		}

	default:
		err = &ErrTypeDiffer{
			etype,
			[]byte{
				ErlTypeSmallInteger,
				ErlTypeInteger,
				ErlTypeSmallBig,
				ErlTypeLargeBig,
			},
		}
	}

	return
}

func reverse(b []byte) []byte {
	size := len(b)
	r := make([]byte, size)

	for i := 0; i < size; i++ {
		r[i] = b[size-i-1]
	}

	return r
}

func termType(r io.Reader) (byte, error) {
	var err error
	b := make([]byte, 1)
	_, err = io.ReadFull(r, b)
	return b[0], err
}
