// Package read implements reading of Erlang external terms.
package read

import (
	"encoding/binary"
	"errors"
	"fmt"
	t "github.com/goerlang/etf/types"
	"io"
	"math"
	"math/big"
)

type ErrTypeDiffer struct {
	Got byte
	Exp []byte
}

type ErrUnknownTerm struct {
	termType byte
}

var be = binary.BigEndian

var (
	ErrFloatScan    = errors.New("read: failed to sscanf float")
	ErrImproperList = errors.New("read: improper list")
	ErrIntTooBig    = errors.New("read: integer too big")
	ErrBadBoolean   = errors.New("read: invalid boolean")
)

func (e *ErrTypeDiffer) Error() string {
	exp := make([]string, len(e.Exp))
	for i, v := range e.Exp {
		exp[i] = fmt.Sprintf("%s(%d)", t.TypeName(v), v)
	}
	return fmt.Sprintf("read: type expected one of %s, got %s(%d)",
		exp, t.TypeName(e.Got), e.Got,
	)
}

func (e *ErrUnknownTerm) Error() string {
	return fmt.Sprintf("read: unknown term type %d", e.termType)
}

func readAtom(r io.Reader, etype byte) (ret t.Atom, err error) {
	switch etype {
	case t.EttAtom:
		// $dLL…
		var size uint16
		if err = binary.Read(r, binary.BigEndian, &size); err == nil {
			b := make([]byte, int(size))
			_, err = io.ReadFull(r, b)
			ret = t.Atom(b)
		}

	case t.EttSmallAtom:
		// $sL…
		var size uint8
		if err = binary.Read(r, binary.BigEndian, &size); err == nil {
			b := make([]byte, int(size))
			_, err = io.ReadFull(r, b)
			ret = t.Atom(b)
		}

	default:
		err = &ErrTypeDiffer{etype, []byte{t.EttAtom, t.EttSmallAtom}}
	}

	return
}

func readBigInt(r io.Reader, etype byte) (ret *big.Int, err error) {
	var size uint32
	var sign byte

	switch etype {
	case t.EttSmallBig:
		// $nAS…
		b := make([]byte, 2)
		if _, err = io.ReadFull(r, b); err == nil {
			size = uint32(b[0])
			sign = b[1]
		}

	case t.EttLargeBig:
		// $oAAAAS…
		b := make([]byte, 5)
		if _, err = io.ReadFull(r, b); err == nil {
			size = binary.BigEndian.Uint32(b[:4])
			sign = b[4]
		}

	default:
		err = &ErrTypeDiffer{etype, []byte{t.EttSmallBig, t.EttLargeBig}}
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

func readBinary(r io.Reader, etype byte) (ret []byte, err error) {
	switch etype {
	case t.EttBinary:
		// $mLLLL…
		var size uint32
		if err = binary.Read(r, binary.BigEndian, &size); err == nil {
			ret = make([]byte, size)
			_, err = io.ReadFull(r, ret)
		}

	case t.EttString:
		// $kLL…
		var size uint16
		if err = binary.Read(r, binary.BigEndian, &size); err == nil {
			ret = make([]byte, size)
			_, err = io.ReadFull(r, ret)
		}

	default:
		err = &ErrTypeDiffer{etype, []byte{t.EttBinary, t.EttString}}
	}

	return
}

func readBool(r io.Reader, etype byte) (ret bool, err error) {
	v, err := readAtom(r, etype)
	if err != nil {
		return
	}

	switch v {
	case t.Atom("true"):
		ret = true

	case t.Atom("false"):
		ret = false

	default:
		err = ErrBadBoolean
	}

	return
}

func readFloat(r io.Reader, etype byte) (ret float64, err error) {
	switch etype {
	case t.EttFloat:
		// $cFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF0
		b := make([]byte, 31)
		if _, err = io.ReadFull(r, b); err == nil {
			var r int
			if r, err = fmt.Sscanf(string(b), "%f", &ret); r != 1 && err == nil {
				err = ErrFloatScan
			}
		}

	case t.EttNewFloat:
		// $FFFFFFFFF
		b := make([]byte, 8)
		if _, err = io.ReadFull(r, b); err == nil {
			ret = math.Float64frombits(be.Uint64(b))
		}

	default:
		err = &ErrTypeDiffer{etype, []byte{t.EttFloat, t.EttNewFloat}}
	}

	return
}

func readInt(r io.Reader, etype byte) (ret int64, err error) {
	switch etype {
	case t.EttSmallInteger:
		// $aI
		var x uint8
		err = binary.Read(r, binary.BigEndian, &x)
		ret = int64(x)

	case t.EttInteger:
		// $bIIII
		var x int32
		err = binary.Read(r, binary.BigEndian, &x)
		ret = int64(x)

	case t.EttSmallBig, t.EttLargeBig:
		var v *big.Int
		if v, err = readBigInt(r, etype); err == nil {
			ret = v.Int64()

			if v.Cmp(big.NewInt(ret)) != 0 {
				err = ErrIntTooBig
			}
		}

	default:
		err = &ErrTypeDiffer{
			etype,
			[]byte{
				t.EttSmallInteger,
				t.EttInteger,
				t.EttSmallBig,
				t.EttLargeBig,
			},
		}
	}

	return
}

func readList(r io.Reader, etype byte) (list t.List, err error) {
	switch etype {
	case t.EttNil:
		list = t.List{}

	case t.EttList:
		// $lLLLL…$j
		var listLen uint32
		if err = binary.Read(r, binary.BigEndian, &listLen); err != nil {
			return
		}

		list = make(t.List, listLen)
		for i := uint32(0); i < listLen; i++ {
			if list[i], err = Term(r); err != nil {
				return
			}
		}

		b := make([]byte, 1)
		_, err = io.ReadFull(r, b)
		if err == nil && b[0] != t.EttNil {
			err = &ErrTypeDiffer{b[0], []byte{t.EttNil}}
		}

	default:
		err = &ErrTypeDiffer{etype, []byte{t.EttNil, t.EttList}}
	}

	return
}

func readPid(r io.Reader, etype byte) (ret t.Pid, err error) {
	switch etype {
	case t.EttPid:
		var b = make([]byte, 9)
		if etype, err = getEtype(r); err != nil {
			return
		} else if ret.Node, err = readAtom(r, etype); err != nil {
			return
		} else if _, err = io.ReadFull(r, b); err != nil {
			return
		}
		ret.Id = binary.BigEndian.Uint32(b[:4])
		ret.Serial = binary.BigEndian.Uint32(b[4:8])
		ret.Creation = b[8]

	default:
		err = &ErrTypeDiffer{etype, []byte{t.EttPid}}
	}

	return
}

func readRef(r io.Reader, etype byte) (ref t.Ref, err error) {
	b := make([]byte, 1)

	switch etype {
	case t.EttNewReference:
		// $rLL…
		var nid uint16
		if err = binary.Read(r, binary.BigEndian, &nid); err != nil {
			return
		} else if etype, err = getEtype(r); err != nil {
			return
		} else if ref.Node, err = readAtom(r, etype); err != nil {
			return
		} else if _, err = io.ReadFull(r, b); err != nil {
			return
		}
		ref.Creation = b[0]

		ref.Id = make([]uint32, nid)
		for i := 0; i < cap(ref.Id); i++ {
			if err = binary.Read(r, binary.BigEndian, &ref.Id[i]); err != nil {
				return
			}
		}

	case t.EttReference:
		// $e…LLLLB
		if etype, err = getEtype(r); err != nil {
			return
		} else if ref.Node, err = readAtom(r, etype); err != nil {
			return
		}
		ref.Id = make([]uint32, 1)
		if err = binary.Read(r, binary.BigEndian, &ref.Id[0]); err != nil {
			return
		} else if _, err = io.ReadFull(r, b); err != nil {
			return
		}
		ref.Creation = b[0]

	default:
		err = &ErrTypeDiffer{
			etype,
			[]byte{
				t.EttReference,
				t.EttNewReference,
			},
		}
	}

	return
}

func readString(r io.Reader, etype byte) (ret string, err error) {
	switch etype {
	case t.EttString, t.EttBinary:
		var b []byte
		b, err = readBinary(r, etype)
		ret = string(b)

	case t.EttList:
		// $lLLLL…$j
		var size uint32
		if err = binary.Read(r, binary.BigEndian, &size); err != nil {
			return
		}

		for i := uint32(0); i < size; i++ {
			if etype, err = getEtype(r); err != nil {
				return
			}

			switch etype {
			case t.EttSmallInteger, t.EttInteger, t.EttSmallBig, t.EttLargeBig:
				var char int64
				if char, err = readInt(r, etype); err != nil {
					return
				}

				ret += string(char)

			default:
				err = &ErrTypeDiffer{
					etype,
					[]byte{
						t.EttSmallInteger,
						t.EttInteger,
						t.EttSmallBig,
						t.EttLargeBig,
					},
				}
				return
			}
		}

		if etype, err = getEtype(r); err == nil && etype != t.EttNil {
			err = ErrImproperList
		}

	case t.EttNil:
		// $j

	default:
		err = &ErrTypeDiffer{
			etype,
			[]byte{
				t.EttString,
				t.EttBinary,
				t.EttList,
				t.EttNil,
			},
		}
	}

	return
}

func Term(r io.Reader) (term t.Term, err error) {
	var etype byte
	if etype, err = getEtype(r); err != nil {
		return nil, err
	}

	switch etype {
	case t.EttAtom, t.EttSmallAtom:
		if term, err = readAtom(r, etype); err != nil {
			return term, err
		} else if term == t.Atom("true") {
			term = true
		} else if term == t.Atom("false") {
			term = false
		}
		return
	case t.EttBinary:
		return readBinary(r, etype)
	case t.EttFloat, t.EttNewFloat:
		return readFloat(r, etype)
	case t.EttSmallInteger, t.EttInteger, t.EttSmallBig, t.EttLargeBig:
		return readInt(r, etype)
	case t.EttString:
		return readString(r, etype)
	case t.EttPid:
		return readPid(r, etype)
	case t.EttReference, t.EttNewReference:
		return readRef(r, etype)
	case t.EttSmallTuple, t.EttLargeTuple:
		return readTuple(r, etype)
	case t.EttNil, t.EttList:
		return readList(r, etype)
		/*
			case t.EttBitBinary:
			case t.EttCachedAtom:
			case t.EttExport:
			case t.EttFun:
			case t.EttList:
			case t.EttNewCache:
			case t.EttNewFun:
			case t.EttPort:
		*/
	}

	return nil, &ErrUnknownTerm{etype}
}

func readTuple(r io.Reader, etype byte) (tuple t.Tuple, err error) {
	var arity int
	switch etype {
	case t.EttSmallTuple:
		// $hA…
		var a uint8
		if err = binary.Read(r, binary.BigEndian, &a); err == nil {
			arity = int(a)
		}

	case t.EttLargeTuple:
		// $iAAAA…
		var a uint32
		if err = binary.Read(r, binary.BigEndian, &a); err == nil {
			arity = int(a)
		}

	default:
		err = &ErrTypeDiffer{etype, []byte{t.EttSmallTuple, t.EttLargeTuple}}
	}

	if err != nil {
		return
	}

	tuple = make(t.Tuple, arity)
	for i := 0; i < arity; i++ {
		if tuple[i], err = Term(r); err != nil {
			break
		}
	}

	return
}

func readUint(r io.Reader, etype byte) (ret uint64, err error) {
	switch etype {
	case t.EttSmallInteger:
		// $aI
		var x uint8
		err = binary.Read(r, binary.BigEndian, &x)
		ret = uint64(x)

	case t.EttInteger:
		// $bIIII
		var x int32
		err = binary.Read(r, binary.BigEndian, &x)
		ret = uint64(x)

	case t.EttSmallBig, t.EttLargeBig:
		var v *big.Int
		if v, err = readBigInt(r, etype); err == nil {
			ret = v.Uint64()

			if v.Cmp(new(big.Int).SetUint64(ret)) != 0 {
				err = ErrIntTooBig
			}
		}

	default:
		err = &ErrTypeDiffer{
			etype,
			[]byte{
				t.EttSmallInteger,
				t.EttInteger,
				t.EttSmallBig,
				t.EttLargeBig,
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

func getEtype(r io.Reader) (byte, error) {
	var err error
	b := make([]byte, 1)
	_, err = io.ReadFull(r, b)
	return b[0], err
}
