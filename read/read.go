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

type ParseReader interface {
	io.Reader
	TermType() byte
}

type termReader struct {
	io.Reader
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

func (tr *termReader) TermType() byte {
	return tr.termType
}

func Atom(r io.Reader) (ret t.Atom, err error) {
	etype, err := termType(r)
	if err != nil {
		return
	}

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
	case t.Atom("true"):
		ret = true

	case t.Atom("false"):
		ret = false

	default:
		err = ErrBadBoolean
	}

	return
}

func Float(r io.Reader) (ret float64, err error) {
	etype, err := termType(r)
	if err != nil {
		return
	}

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

func Int(r io.Reader) (ret int64, err error) {
	etype, err := termType(r)
	if err == nil {
		ret, err = getInt(etype, r)
	}

	return
}

func List(r io.Reader) (list t.List, err error) {
	etype, err := termType(r)
	if err != nil {
		return
	}

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

func Pid(r io.Reader) (ret t.Pid, err error) {
	etype, err := termType(r)
	if err != nil {
		return
	}

	switch etype {
	case t.EttPid:
		var a t.Atom
		var b = make([]byte, 9)
		if a, err = Atom(r); err != nil {
			return
		} else if _, err = io.ReadFull(r, b); err != nil {
			return
		}
		ret.Node = t.Node(a)
		ret.Id = binary.BigEndian.Uint32(b[:4])
		ret.Serial = binary.BigEndian.Uint32(b[4:8])
		ret.Creation = b[8]

	default:
		err = &ErrTypeDiffer{etype, []byte{t.EttPid}}
	}

	return
}

func Ref(r io.Reader) (ref t.Ref, err error) {
	etype, err := termType(r)
	if err != nil {
		return
	}

	var node t.Atom
	b := make([]byte, 1)

	switch etype {
	case t.EttNewReference:
		// $rLL…
		var nid uint16
		if err = binary.Read(r, binary.BigEndian, &nid); err != nil {
			return
		}
		if node, err = Atom(r); err != nil {
			return
		}
		if _, err = io.ReadFull(r, b); err != nil {
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
		if node, err = Atom(r); err != nil {
			return
		}
		ref.Id = make([]uint32, 1)
		if err = binary.Read(r, binary.BigEndian, &ref.Id[0]); err != nil {
			return
		}
		if _, err = io.ReadFull(r, b); err != nil {
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

	ref.Node = t.Node(node)

	return
}

func String(r io.Reader) (ret string, err error) {
	etype, err := termType(r)
	if err != nil {
		return
	}

	switch etype {
	case t.EttString, t.EttBinary:
		var b []byte
		b, err = getBinary(etype, r)
		ret = string(b)

	case t.EttList:
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
			case t.EttSmallInteger, t.EttInteger, t.EttSmallBig, t.EttLargeBig:
				var char int64
				if char, err = getInt(etype, r); err != nil {
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

		if _, err = io.ReadFull(r, b); err == nil && b[0] != t.EttNil {
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
	b := make([]byte, 1)
	if _, err = io.ReadFull(r, b); err != nil {
		return nil, err
	}

	tr := termReader{
		r,
		b[0],
	}

	switch b[0] {
	case t.EttAtom, t.EttSmallAtom:
		if term, err = Atom(&tr); err != nil {
			return term, err
		} else if term == t.Atom("true") {
			term = true
		} else if term == t.Atom("false") {
			term = false
		}
		return
	case t.EttBinary:
		return Binary(&tr)
	case t.EttFloat, t.EttNewFloat:
		return Float(&tr)
	case t.EttSmallInteger, t.EttInteger, t.EttSmallBig, t.EttLargeBig:
		return Int(&tr)
	case t.EttString:
		return String(&tr)
	case t.EttPid:
		return Pid(&tr)
	case t.EttReference, t.EttNewReference:
		return Ref(&tr)
	case t.EttSmallTuple, t.EttLargeTuple:
		return Tuple(&tr)
	case t.EttNil, t.EttList:
		return List(&tr)
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

	return nil, &ErrUnknownTerm{b[0]}
}

func Tuple(r io.Reader) (tuple t.Tuple, err error) {
	etype, err := termType(r)
	if err != nil {
		return
	}

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

func Uint(r io.Reader) (ret uint64, err error) {
	etype, err := termType(r)
	if err == nil {
		ret, err = getUint(etype, r)
	}

	return
}

func getBigInt(etype byte, r io.Reader) (ret *big.Int, err error) {
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

func getBinary(etype byte, r io.Reader) (ret []byte, err error) {
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

func getInt(etype byte, r io.Reader) (ret int64, err error) {
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
				t.EttSmallInteger,
				t.EttInteger,
				t.EttSmallBig,
				t.EttLargeBig,
			},
		}
	}

	return
}

func getUint(etype byte, r io.Reader) (ret uint64, err error) {
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
		if v, err = getBigInt(etype, r); err == nil {
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

func termType(r io.Reader) (byte, error) {
	if rt, ok := r.(ParseReader); ok {
		return rt.TermType(), nil
	}

	var err error
	b := make([]byte, 1)
	_, err = io.ReadFull(r, b)
	return b[0], err
}
