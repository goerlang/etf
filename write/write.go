// Package write implements writing basic Go types as external Erlang terms.
package write

import (
	"bytes"
	"fmt"
	t "github.com/goerlang/etf/types"
	"io"
	"math"
	"math/big"
	"reflect"
)

type ErrUnknownType struct {
	t reflect.Type
}

func (e *ErrUnknownType) Error() string {
	return fmt.Sprintf("write: can't encode type \"%s\"", e.t.Name())
}

func Atom(w io.Writer, atom t.Atom) (err error) {
	switch size := len(atom); {
	case size <= 0xff:
		// $sL…
		if _, err = w.Write([]byte{t.EttSmallAtom, byte(size)}); err == nil {
			_, err = w.Write([]byte(atom))
		}

	case size <= 0xffff:
		// $dLL…
		_, err = w.Write([]byte{byte(t.EttAtom), byte(size >> 8), byte(size)})
		if err == nil {
			_, err = w.Write([]byte(atom))
		}

	default:
		err = fmt.Errorf("atom is too big (%d bytes)", size)
	}

	return
}

func BigInt(w io.Writer, x *big.Int) (err error) {
	sign := 0
	if x.Sign() < 0 {
		sign = 1
	}

	bytes := reverse(new(big.Int).Abs(x).Bytes())

	switch size := len(bytes); {
	case size <= 0xff:
		// $nAS…
		_, err = w.Write([]byte{t.EttSmallBig, byte(size), byte(sign)})

	case int(uint32(size)) == size:
		// $oAAAAS…
		_, err = w.Write([]byte{
			t.EttLargeBig,
			byte(size >> 24), byte(size >> 16), byte(size >> 8), byte(size),
			byte(sign),
		})

	default:
		err = fmt.Errorf("bad big int size (%d)", size)
	}

	if err == nil {
		_, err = w.Write(bytes)
	}

	return
}

func Binary(w io.Writer, bytes []byte) (err error) {
	switch size := len(bytes); {
	case int(uint32(size)) == size:
		// $mLLLL…
		data := []byte{
			t.EttBinary,
			byte(size >> 24), byte(size >> 16), byte(size >> 8), byte(size),
		}
		if _, err = w.Write(data); err == nil {
			_, err = w.Write(bytes)
		}

	default:
		err = fmt.Errorf("bad binary size (%d)", size)
	}

	return
}

func Bool(w io.Writer, b bool) (err error) {
	if b {
		err = Atom(w, t.Atom("true"))
	} else {
		err = Atom(w, t.Atom("false"))
	}

	return
}

func Float(w io.Writer, f float64) (err error) {
	if _, err = w.Write([]byte{t.EttNewFloat}); err == nil {
		fb := math.Float64bits(f)
		_, err = w.Write([]byte{
			byte(fb >> 56), byte(fb >> 48), byte(fb >> 40), byte(fb >> 32),
			byte(fb >> 24), byte(fb >> 16), byte(fb >> 8), byte(fb),
		})
	}
	return
}

func Int(w io.Writer, x int64) (err error) {
	switch {
	case x >= 0 && x <= math.MaxUint8:
		// $aI
		_, err = w.Write([]byte{t.EttSmallInteger, byte(x)})

	case x >= math.MinInt32 && x <= math.MaxInt32:
		// $bIIII
		_, err = w.Write([]byte{
			t.EttInteger,
			byte(x >> 24), byte(x >> 16), byte(x >> 8), byte(x),
		})

	default:
		err = BigInt(w, big.NewInt(x))
	}

	return
}

func Uint(w io.Writer, x uint64) (err error) {
	switch {
	case x <= math.MaxUint8:
		// $aI
		_, err = w.Write([]byte{t.EttSmallInteger, byte(x)})

	case x <= math.MaxInt32:
		// $bIIII
		_, err = w.Write([]byte{
			t.EttInteger,
			byte(x >> 24), byte(x >> 16), byte(x >> 8), byte(x),
		})

	default:
		err = BigInt(w, new(big.Int).SetUint64(x))
	}

	return
}


func Pid(w io.Writer, p t.Pid) (err error) {
	if _, err = w.Write([]byte{t.EttPid}); err != nil {
		return
	} else if err = Atom(w, t.Atom(p.Node)); err != nil {
		return
	}

	_, err = w.Write([]byte{
		0, 0, byte(p.Id >> 8), byte(p.Id),
		byte(p.Serial >> 24),
		byte(p.Serial >> 16),
		byte(p.Serial >> 8),
		byte(p.Serial),
		p.Creation,
	})

	return
}

func String(w io.Writer, s string) (err error) {
	switch size := len(s); {
	case size <= 0xffff:
		// $kLL…
		_, err = w.Write([]byte{t.EttString, byte(size >> 8), byte(size)})
		if err == nil {
			_, err = w.Write([]byte(s))
		}

	default:
		err = fmt.Errorf("string is too big (%d bytes)", size)
	}

	return
}

func List(w io.Writer, l interface{}) (err error) {
	rv := reflect.ValueOf(l)
	n := rv.Len()
	_, err = w.Write([]byte{
		t.EttList,
		byte(n >> 24),
		byte(n >> 16),
		byte(n >> 8),
		byte(n),
	})

	if err != nil {
		return
	}

	for i := 0; i < n; i++ {
		v := rv.Index(i).Interface()
		if err = Term(w, v); err != nil {
			return
		}
	}

	_, err = w.Write([]byte{t.EttNil})

	return
}

func Record(w io.Writer, r interface{}) (err error) {
	rv := reflect.ValueOf(r)
	n := rv.NumField()
	buf := new(bytes.Buffer)
	arity := 0

	for i := 0; i < n; i++ {
		if f := rv.Field(i); f.CanInterface() {
			if err = Term(buf, f.Interface()); err != nil {
				return
			}
			arity++
		}
	}

	if arity <= math.MaxUint8 {
		_, err = w.Write([]byte{t.EttSmallTuple, byte(arity)})
	} else {
		_, err = w.Write([]byte{
			t.EttLargeTuple,
			byte(arity >> 24),
			byte(arity >> 16),
			byte(arity >> 8),
			byte(arity),
		})
	}

	if err == nil {
		_, err = buf.WriteTo(w)
	}

	return
}

func Tuple(w io.Writer, tuple t.Tuple) (err error) {
	n := len(tuple)
	if n <= math.MaxUint8 {
		_, err = w.Write([]byte{t.EttSmallTuple, byte(n)})
	} else {
		_, err = w.Write([]byte{
			t.EttLargeTuple,
			byte(n >> 24),
			byte(n >> 16),
			byte(n >> 8),
			byte(n),
		})
	}

	if err != nil {
		return
	}

	for _, v := range tuple {
		if err = Term(w, v); err != nil {
			return
		}
	}

	return
}

func Term(w io.Writer, term t.Term) (err error) {
	switch v := term.(type) {
	case bool:
		return Bool(w, v)
	case int8, int16, int32, int64, int:
		return Int(w, reflect.ValueOf(term).Int())
	case uint8, uint16, uint32, uint64, uintptr, uint:
		return Uint(w, reflect.ValueOf(term).Uint())
	case string:
		return String(w, v)
	case []byte:
		return Binary(w, v)
	case t.Atom:
		return Atom(w, v)
	case float64:
		return Float(w, v)
	case float32:
		return Float(w, float64(v))
	case t.Pid:
		return Pid(w, v)
	case t.Tuple:
		return Tuple(w, v)
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Struct:
			return Record(w, term)
		case reflect.Array, reflect.Slice:
			return List(w, term)
		case reflect.Ptr:
			return Term(w, rv.Elem())
		//case reflect.Map // FIXME
		default:
			err = &ErrUnknownType{rv.Type()}
		}
	}

	return
}

func reverse(b []byte) []byte {
	size := len(b)
	hsize := size >> 1

	for i := 0; i < hsize; i++ {
		b[i], b[size-i-1] = b[size-i-1], b[i]
	}

	return b
}
