// Package read implements reading of Erlang external terms.
package read

import (
	"bytes"
	"encoding/binary"
	"fmt"
	t "github.com/goerlang/etf/types"
	"io"
	"math"
	"math/big"
)

type ErrUnknownTerm struct {
	termType byte
}

var (
	ErrFloatScan = fmt.Errorf("read: failed to sscanf float")
	be           = binary.BigEndian
	bTrue        = []byte("true")
	bFalse       = []byte("false")
)

func (e *ErrUnknownTerm) Error() string {
	return fmt.Sprintf("read: unknown term type %d", e.termType)
}

func atom(b []byte) interface{} {
	if bytes.Compare(b, bTrue) == 0 {
		return true
	} else if bytes.Compare(b, bFalse) == 0 {
		return false
	}
	return t.Atom(b)
}

func bigInt(r io.Reader, b []byte, sign byte) (interface{}, error) {
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}

	size := len(b)
	hsize := size >> 1
	for i := 0; i < hsize; i++ {
		b[i], b[size-i-1] = b[size-i-1], b[i]
	}

	v := new(big.Int).SetBytes(b)
	if sign != 0 {
		v = v.Neg(v)
	}

	// try int and int64
	v64 := v.Int64()
	if x := int(v64); v.Cmp(big.NewInt(int64(x))) == 0 {
		return x, nil
	} else if v.Cmp(big.NewInt(v64)) == 0 {
		return v64, nil
	}

	return v, nil
}

func Term(r io.Reader) (term t.Term, err error) {
	var etype byte
	if etype, err = ruint8(r); err != nil {
		return nil, err
	}
	var b []byte

	switch etype {
	case t.EttAtom, t.EttAtomUTF8:
		// $dLL… | $vLL…
		if b, err = buint16(r); err == nil {
			_, err = io.ReadFull(r, b)
			term = atom(b)
		}

	case t.EttSmallAtom, t.EttSmallAtomUTF8:
		// $sL…, $wL…
		if b, err = buint8(r); err == nil {
			_, err = io.ReadFull(r, b)
			term = atom(b)
		}

	case t.EttBinary:
		// $mLLLL…
		if b, err = buint32(r); err == nil {
			_, err = io.ReadFull(r, b)
			term = b
		}

	case t.EttString:
		// $kLL…
		if b, err = buint16(r); err == nil {
			_, err = io.ReadFull(r, b)
			term = string(b)
		}

	case t.EttFloat:
		// $cFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF0
		b = make([]byte, 31)
		if _, err = io.ReadFull(r, b); err != nil {
			return
		}
		var r int
		var f float64
		if r, err = fmt.Sscanf(string(b), "%f", &f); r != 1 && err == nil {
			err = ErrFloatScan
		}
		term = f

	case t.EttNewFloat:
		// $FFFFFFFFF
		b = make([]byte, 8)
		if _, err = io.ReadFull(r, b); err == nil {
			term = math.Float64frombits(be.Uint64(b))
		}

	case t.EttSmallInteger:
		// $aI
		var x uint8
		x, err = ruint8(r)
		term = int(x)

	case t.EttInteger:
		// $bIIII
		var x int32
		err = binary.Read(r, be, &x)
		term = int(x)

	case t.EttSmallBig:
		// $nAS…
		b = make([]byte, 2)
		if _, err = io.ReadFull(r, b); err != nil {
			break
		}
		sign := b[1]
		b = make([]byte, b[0])
		term, err = bigInt(r, b, sign)

	case t.EttLargeBig:
		// $oAAAAS…
		b = make([]byte, 5)
		if _, err = io.ReadFull(r, b); err != nil {
			break
		}
		sign := b[4]
		b = make([]byte, be.Uint32(b[:4]))
		term, err = bigInt(r, b, sign)

	case t.EttNil:
		// $j
		term = t.List{}

	case t.EttPid:
		var node interface{}
		var pid t.Pid
		b = make([]byte, 9)
		if node, err = Term(r); err != nil {
			return
		} else if _, err = io.ReadFull(r, b); err != nil {
			return
		}
		pid.Node = node.(t.Atom)
		pid.Id = be.Uint32(b[:4])
		pid.Serial = be.Uint32(b[4:8])
		pid.Creation = b[8]
		term = pid

	case t.EttNewReference:
		// $rLL…
		var ref t.Ref
		var node interface{}
		var nid uint16
		if nid, err = ruint16(r); err != nil {
			return
		} else if node, err = Term(r); err != nil {
			return
		} else if ref.Creation, err = ruint8(r); err != nil {
			return
		}
		ref.Node = node.(t.Atom)
		ref.Id = make([]uint32, nid)
		for i := 0; i < cap(ref.Id); i++ {
			if ref.Id[i], err = ruint32(r); err != nil {
				return
			}
		}
		term = ref

	case t.EttReference:
		// $e…LLLLB
		var ref t.Ref
		var node interface{}
		if node, err = Term(r); err != nil {
			return
		}
		ref.Node = node.(t.Atom)
		ref.Id = make([]uint32, 1)
		if ref.Id[0], err = ruint32(r); err != nil {
			return
		} else if _, err = io.ReadFull(r, b); err != nil {
			return
		}
		ref.Creation = b[0]
		term = ref

	case t.EttSmallTuple:
		// $hA…
		var arity uint8
		if arity, err = ruint8(r); err != nil {
			break
		}
		tuple := make(t.Tuple, arity)
		for i := 0; i < cap(tuple); i++ {
			if tuple[i], err = Term(r); err != nil {
				break
			}
		}
		term = tuple

	case t.EttLargeTuple:
		// $iAAAA…
		var arity uint32
		if arity, err = ruint32(r); err != nil {
			break
		}
		tuple := make(t.Tuple, arity)
		for i := 0; i < cap(tuple); i++ {
			if tuple[i], err = Term(r); err != nil {
				break
			}
		}
		term = tuple

	case t.EttList:
		// $lLLLL…$j
		var n uint32
		if n, err = ruint32(r); err != nil {
			return
		}

		list := make(t.List, n+1)
		for i := 0; i < cap(list); i++ {
			if list[i], err = Term(r); err != nil {
				return
			}
		}

		switch list[n].(type) {
		case t.List:
			// proper list, remove nil element
			list = list[:n]
		}
		term = list

	case t.EttBitBinary:
		// $MLLLLB…
		var length uint32
		var bits uint8
		if length, err = ruint32(r); err != nil {
			break
		} else if bits, err = ruint8(r); err != nil {
			break
		}
		b := make([]byte, length)
		_, err = io.ReadFull(r, b)
		b[len(b)-1] = b[len(b)-1] >> (8 - bits)
		term = b

	case t.EttExport:
		// $qM…F…A
		var m, f interface{}
		var a uint8
		if m, err = Term(r); err != nil {
			break
		} else if f, err = Term(r); err != nil {
			break
		} else if a, err = ruint8(r); err != nil {
			break
		}

		term = t.Export{m.(t.Atom), f.(t.Atom), a}

		/*
			case t.EttCachedAtom:
			case t.EttFun:
			case t.EttNewCache:
			case t.EttNewFun:
			case t.EttPort:
		*/
	default:
		err = &ErrUnknownTerm{etype}
	}

	return
}

func ruint8(r io.Reader) (uint8, error) {
	b := []byte{0}
	_, err := io.ReadFull(r, b)
	return b[0], err
}

func ruint16(r io.Reader) (uint16, error) {
	b := []byte{0, 0}
	_, err := io.ReadFull(r, b)
	return be.Uint16(b), err
}

func ruint32(r io.Reader) (uint32, error) {
	b := []byte{0, 0, 0, 0}
	_, err := io.ReadFull(r, b)
	return be.Uint32(b), err
}

func buint8(r io.Reader) ([]byte, error) {
	size, err := ruint8(r)
	return make([]byte, size), err
}

func buint16(r io.Reader) ([]byte, error) {
	size, err := ruint16(r)
	return make([]byte, size), err
}

func buint32(r io.Reader) ([]byte, error) {
	size, err := ruint32(r)
	return make([]byte, size), err
}
