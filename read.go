package etf

import (
	"bytes"
	"encoding/binary"
	"fmt"
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

func (c *Context) ReadDist(r io.Reader) (err error) {
	b := make([]byte, 1)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return
	}

	if b[0] != EtDist {
		err = fmt.Errorf("Not dist header: %d", b[0])
		return
	}

	_, err = io.ReadFull(r, b)
	if err != nil {
		return
	}

	refsNum := int(b[0])
	if refsNum > 0 {
		b = make([]byte, (refsNum/2)+1)
		_, err = io.ReadFull(r, b)
		if err != nil {
			return
		}

		flags := make([]cacheFlag, refsNum)
		longAtoms := false

		for i := 0; i < (refsNum + 1); i++ {
			var v byte
			if (i & 0x01) == 0 {
				v = b[i/2] & 0x0F
			} else {
				v = (b[i/2] >> 4) & 0x0F
			}
			if i < refsNum {
				flags[i] = cacheFlag{(v & 0x08) == 0x08, v & 0x07}
			} else {
				longAtoms = (v & 0x01) == 0x01
			}
		}

		var headLen int
		if longAtoms {
			headLen = 1 + 2
		} else {
			headLen = 1 + 1
		}
		currentAtomCache := make([]*string, len(flags))
		i := 0

		for _, _ = range flags {
			if flags[i].isNew {
				b = make([]byte, headLen)
				_, err = io.ReadFull(r, b)
				if err != nil {
					return
				}
				intRef := uint8(b[0])

				var atomLen uint
				if longAtoms {
					atomLen = uint(binary.BigEndian.Uint16(b[1:3]))
				} else {
					atomLen = uint(b[1])
				}
				b = make([]byte, atomLen)
				_, err = io.ReadFull(r, b)
				if err != nil {
					return
				}
				strText := string(b)
				currentAtomCache[i] = &strText

				cIdx := ((uint16(flags[i].segmentIdx) << 8) | uint16(intRef))
				c.atomCache[cIdx] = &strText
			} else {
				b = make([]byte, 1)
				_, err = io.ReadFull(r, b)
				if err != nil {
					return
				}
				intRef := uint8(b[0])
				cIdx := ((uint16(flags[i].segmentIdx) << 8) | uint16(intRef))
				currentAtomCache[i] = c.atomCache[cIdx]
			}
			i++
		}

		c.currentCache = currentAtomCache
	}
	return
}

func (c *Context) Read(r io.Reader) (term Term, err error) {
	var etype byte
	if etype, err = ruint8(r); err != nil {
		return nil, err
	}
	var b []byte

	switch etype {
	case ettAtom, ettAtomUTF8:
		// $dLL… | $vLL…
		if b, err = buint16(r); err == nil {
			_, err = io.ReadFull(r, b)
			term = newAtom(b)
		}

	case ettSmallAtom, ettSmallAtomUTF8:
		// $sL…, $wL…
		if b, err = buint8(r); err == nil {
			_, err = io.ReadFull(r, b)
			term = newAtom(b)
		}

	case ettBinary:
		// $mLLLL…
		if b, err = buint32(r); err == nil {
			_, err = io.ReadFull(r, b)
			term = b
		}

	case ettString:
		// $kLL…
		if b, err = buint16(r); err == nil {
			_, err = io.ReadFull(r, b)
			term = string(b)
		}

	case ettFloat:
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

	case ettNewFloat:
		// $FFFFFFFFF
		b = make([]byte, 8)
		if _, err = io.ReadFull(r, b); err == nil {
			term = math.Float64frombits(be.Uint64(b))
		}

	case ettSmallInteger:
		// $aI
		var x uint8
		x, err = ruint8(r)
		term = int(x)

	case ettInteger:
		// $bIIII
		var x int32
		err = binary.Read(r, be, &x)
		term = int(x)

	case ettSmallBig:
		// $nAS…
		b = make([]byte, 2)
		if _, err = io.ReadFull(r, b); err != nil {
			break
		}
		sign := b[1]
		b = make([]byte, b[0])
		term, err = readBigInt(r, b, sign)

	case ettLargeBig:
		// $oAAAAS…
		b = make([]byte, 5)
		if _, err = io.ReadFull(r, b); err != nil {
			break
		}
		sign := b[4]
		b = make([]byte, be.Uint32(b[:4]))
		term, err = readBigInt(r, b, sign)

	case ettNil:
		// $j
		term = List{}

	case ettPid:
		var node interface{}
		var pid Pid
		b = make([]byte, 9)
		if node, err = c.Read(r); err != nil {
			return
		} else if _, err = io.ReadFull(r, b); err != nil {
			return
		}
		pid.Node = node.(Atom)
		pid.Id = be.Uint32(b[:4])
		pid.Serial = be.Uint32(b[4:8])
		pid.Creation = b[8]
		term = pid

	case ettNewRef:
		// $rLL…
		var ref Ref
		var node interface{}
		var nid uint16
		if nid, err = ruint16(r); err != nil {
			return
		} else if node, err = c.Read(r); err != nil {
			return
		} else if ref.Creation, err = ruint8(r); err != nil {
			return
		}
		ref.Node = node.(Atom)
		ref.Id = make([]uint32, nid)
		for i := 0; i < cap(ref.Id); i++ {
			if ref.Id[i], err = ruint32(r); err != nil {
				return
			}
		}
		term = ref

	case ettRef:
		// $e…LLLLB
		var ref Ref
		var node interface{}
		if node, err = c.Read(r); err != nil {
			return
		}
		ref.Node = node.(Atom)
		ref.Id = make([]uint32, 1)
		if ref.Id[0], err = ruint32(r); err != nil {
			return
		} else if _, err = io.ReadFull(r, b); err != nil {
			return
		}
		ref.Creation = b[0]
		term = ref

	case ettSmallTuple:
		// $hA…
		var arity uint8
		if arity, err = ruint8(r); err != nil {
			break
		}
		tuple := make(Tuple, arity)
		for i := 0; i < cap(tuple); i++ {
			if tuple[i], err = c.Read(r); err != nil {
				break
			}
		}
		term = tuple

	case ettLargeTuple:
		// $iAAAA…
		var arity uint32
		if arity, err = ruint32(r); err != nil {
			break
		}
		tuple := make(Tuple, arity)
		for i := 0; i < cap(tuple); i++ {
			if tuple[i], err = c.Read(r); err != nil {
				break
			}
		}
		term = tuple

	case ettList:
		// $lLLLL…$j
		var n uint32
		if n, err = ruint32(r); err != nil {
			return
		}

		list := make(List, n+1)
		for i := 0; i < cap(list); i++ {
			if list[i], err = c.Read(r); err != nil {
				return
			}
		}

		switch list[n].(type) {
		case List:
			// proper list, remove nil element
			list = list[:n]
		}
		term = list

	case ettBitBinary:
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

	case ettExport:
		// $qM…F…A
		var m, f interface{}
		var a uint8
		if m, err = c.Read(r); err != nil {
			break
		} else if f, err = c.Read(r); err != nil {
			break
		} else if a, err = ruint8(r); err != nil {
			break
		}

		term = Export{m.(Atom), f.(Atom), a}

	case ettNewFun:
		// $pSSSSAUUUUUUUUUUUUUUUUIIIIFFFFM…i…u…P…[V…]
		var f Function
		ruint32(r)
		f.Arity, _ = ruint8(r)
		io.ReadFull(r, f.Unique[:])
		f.Index, _ = ruint32(r)
		f.Free, _ = ruint32(r)
		m, _ := c.Read(r)
		oldi, _ := c.Read(r)
		oldu, _ := c.Read(r)
		pid, _ := c.Read(r)

		f.FreeVars = make([]Term, f.Free)
		for i := 0; i < cap(f.FreeVars); i++ {
			if f.FreeVars[i], err = c.Read(r); err != nil {
				break
			}
		}

		f.Module = m.(Atom)
		f.OldIndex = uint32(oldi.(int))
		f.OldUnique = uint32(oldu.(int))
		f.Pid = pid.(Pid)
		term = f

	case ettFun:
		// $uFFFFP…M…i…u…[V…]
		var f Function
		f.Free, _ = ruint32(r)
		pid, _ := c.Read(r)
		m, _ := c.Read(r)
		oldi, _ := c.Read(r)
		oldu, _ := c.Read(r)

		f.FreeVars = make([]Term, f.Free)
		for i := 0; i < cap(f.FreeVars); i++ {
			if f.FreeVars[i], err = c.Read(r); err != nil {
				break
			}
		}

		f.Module = m.(Atom)
		f.OldIndex = uint32(oldi.(int))
		f.OldUnique = uint32(oldu.(int))
		f.Pid = pid.(Pid)
		term = f

	case ettPort:
		// $fA…IIIIC
		var p Port
		a, _ := c.Read(r)
		p.Node = a.(Atom)
		p.Id, _ = ruint32(r)
		p.Creation, err = ruint8(r)
		term = p

	case ettCacheRef:
		b = make([]byte, 1)
		if _, err = io.ReadFull(r, b); err != nil {
			break
		}
		term = Atom(*c.currentCache[b[0]])

	default:
		err = &ErrUnknownTerm{etype}
	}

	return
}

func (e *ErrUnknownTerm) Error() string {
	return fmt.Sprintf("read: unknown term type %d", e.termType)
}

func newAtom(b []byte) interface{} {
	if bytes.Compare(b, bTrue) == 0 {
		return true
	} else if bytes.Compare(b, bFalse) == 0 {
		return false
	}
	return Atom(b)
}

func readBigInt(r io.Reader, b []byte, sign byte) (interface{}, error) {
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
