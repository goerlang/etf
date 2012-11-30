package etf

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"
)

// writeBE
func writeBE(w io.Writer, data ...interface{}) error {
	for _, v := range data {
		err := binary.Write(w, be, v)

		if err != nil {
			return err
		}
	}

	return nil
}

// writeAtom
func writeAtom(w io.Writer, a Atom) (err error) {
	switch size := len(a); {
	case size <= 0xff:
		// $sL…
		err = writeBE(w, be, byte(erlSmallAtom), byte(size), []byte(string(a)))

	case size <= 0xffff:
		// $dLL…
		err = writeBE(w, be, byte(erlAtom), uint16(size), []byte(string(a)))

	default:
		err = EncodeError{fmt.Sprintf("atom is too big (%d bytes)", size)}
	}

	return
}

// writeBigInt
func writeBigInt(w io.Writer, x *big.Int) (err error) {
	sign := 0
	if x.Sign() < 0 {
		sign = 1
	}

	bytes := reverseBytes(x.Abs(x).Bytes())

	switch size := len(bytes); {
	case size <= 0xff:
		// $nAS…
		err = writeBE(w, be, byte(erlSmallBig), byte(size), byte(sign), bytes)

	case int(uint32(size)) == size:
		// $oAAAAS…
		err = writeBE(w, be, byte(erlLargeBig), uint32(size), byte(sign), bytes)

	default:
		err = EncodeError{fmt.Sprintf("bad big int size (%d)", size)}
	}

	return
}

// writeBinary
func writeBinary(w io.Writer, bytes []byte) (err error) {
	switch size := len(bytes); {
	case int(uint32(size)) == size:
		// $mLLLL…
		err = writeBE(w, be, byte(erlBinary), uint32(len(bytes)), bytes)

	default:
		err = EncodeError{fmt.Sprintf("bad binary size (%d)", size)}
	}

	return
}

// writeBool
func writeBool(w io.Writer, b bool) (err error) {
	switch b {
	case true:
		err = writeAtom(w, Atom("true"))

	case false:
		err = writeAtom(w, Atom("false"))
	}

	return
}

// writeFloat64
func writeFloat64(w io.Writer, f float64) error {
	return writeBE(w, be, byte(erlNewFloat), math.Float64bits(f))
}

// writeInt64
func writeInt64(w io.Writer, x int64) (err error) {
	switch {
	case x >= 0 && x <= 0xff:
		// $aI
		err = writeBE(w, be, byte(erlSmallInteger), byte(x))

	case int64(int32(x)) == x:
		// $bIIII
		err = writeBE(w, be, byte(erlInteger), int32(x))

	default:
		err = writeBigInt(w, big.NewInt(x))
	}

	return
}

// writeString
func writeString(w io.Writer, s string) (err error) {
	switch size := len(s); {
	case size <= 0xffff:
		// $kLL…
		err = writeBE(w, byte(erlString), uint16(size), []byte(s))

	default:
		err = EncodeError{fmt.Sprintf("string is too big (%d bytes)", size)}
	}

	return
}
