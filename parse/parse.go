package parse

import (
	"encoding/binary"
	"fmt"
	. "github.com/goerlang/etf/types"
	"math"
	"math/big"
)

var be = binary.BigEndian

type StructuralError struct {
	Msg string
}

type SyntaxError struct {
	Msg string
}

func (err StructuralError) Error() string {
	return "etf: structural error: " + err.Msg
}

func (err SyntaxError) Error() string {
	return "etf: syntax error: " + err.Msg
}

func Atom(b []byte) (ret ErlAtom, size int, err error) {
	switch ErlType(b[0]) {
	case ErlTypeAtom:
		// $dLL…
		if len(b) >= 3 {
			size = 3 + int(be.Uint16(b[1:3]))

			if len(b) >= size {
				ret = ErlAtom(b[3:size])
			} else {
				err = StructuralError{"wrong atom size"}
			}
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid atom length (%d)", len(b)),
			}
		}

	case ErlTypeSmallAtom:
		// $sL…
		if len(b) >= 2 {
			size = 2 + int(b[1])

			if len(b) >= size {
				ret = ErlAtom(b[2:size])
			} else {
				err = StructuralError{"wrong atom size"}
			}
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid small atom length (%d)", len(b)),
			}
		}

	default:
		err = SyntaxError{"not an atom"}
	}

	return
}

func BigInt(b []byte) (ret *big.Int, size int, err error) {
	switch ErlType(b[0]) {
	case ErlTypeSmallBig:
		// $nAS…
		if len(b) >= 3 && len(b)-3 >= int(b[1]) {
			size = 3 + int(b[1])
			b2 := reverseBytes(b[3 : 3+int(b[1])])
			ret = new(big.Int).SetBytes(b2)

			if b[2] != 0 {
				ret = ret.Neg(ret)
			}
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid small big integer length (%d)", len(b)),
			}
		}

	case ErlTypeLargeBig:
		// $oAAAAS…
		if len(b) < 6 {
			err = StructuralError{
				fmt.Sprintf("invalid large big integer length (%d)", len(b)),
			}
			break
		}

		ulength := be.Uint32(b[1:5])
		length := int(ulength)
		if uint32(length) != ulength {
			err = fmt.Errorf("ErlTypeLargeBig size 0x%x overflows int type", ulength)
			break
		}

		size = 6 + length

		if len(b) >= size {
			b2 := reverseBytes(b[6 : 6+length])
			ret = new(big.Int).SetBytes(b2)

			if b[5] != 0 {
				ret = ret.Neg(ret)
			}
		} else {
			err = StructuralError{"invalid large bigInt"}
		}

	default:
		err = SyntaxError{"not a big integer"}
	}

	return
}

func Binary(b []byte) (ret []byte, size int, err error) {
	var s int

	switch t := ErlType(b[0]); t {
	case ErlTypeBinary:
		// $mLLLL…
		s = 5

	case ErlTypeString:
		// $kLL…
		s = 3

	default:
		err = SyntaxError{"not a binary"}
		return
	}

	if len(b) < s {
		err = StructuralError{
			fmt.Sprintf("invalid binary length (%d)", len(b)),
		}
		return
	}

	usize := uint(s)
	if s == 5 {
		usize += uint(be.Uint32(b[1:s]))
	} else {
		usize += uint(be.Uint16(b[1:s]))
	}

	size = int(usize)
	if uint(size) != usize {
		err = fmt.Errorf("erlBinary/erlString size 0x%x overflows int type", usize)
		return
	}

	if len(b) >= size {
		ret = b[s:size]
	} else {
		err = StructuralError{
			fmt.Sprintf("invalid binary size (%d), len=%d", size, len(b)),
		}
	}

	return
}

func Bool(b []byte) (ret bool, size int, err error) {
	var v ErlAtom

	v, size, err = Atom(b)

	if err == nil {
		switch v {
		case ErlAtom("true"):
			ret = true
			return

		case ErlAtom("false"):
			ret = false
			return
		}

		err = SyntaxError{"not a boolean"}
	}

	return
}

func Float64(b []byte) (ret float64, size int, err error) {
	switch ErlType(b[0]) {
	case ErlTypeFloat:
		// $cFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF0
		if len(b) >= 32 {
			var r int
			r, err = fmt.Sscanf(string(b[1:32]), "%f", &ret)
			size += 32

			if r != 1 || err != nil {
				err = StructuralError{"failed to scan float"}
			}
		} else {
			err = StructuralError{fmt.Sprintf("invalid float length (%d)", len(b))}
		}

	case ErlTypeNewFloat:
		// $FFFFFFFFF
		if len(b) >= 9 {
			ret = math.Float64frombits(be.Uint64(b[1:9]))
			size += 9
		} else {
			err = StructuralError{fmt.Sprintf("invalid float length (%d)", len(b))}
		}

	default:
		err = SyntaxError{"not a float"}
	}

	return
}

func Int64(b []byte) (ret int64, size int, err error) {
	switch ErlType(b[0]) {
	case ErlTypeSmallInteger:
		// $aI
		if len(b) >= 2 {
			return int64(b[1]), 2, nil
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid small integer length (%d)", len(b)),
			}
		}

	case ErlTypeInteger:
		// $bIIII
		if len(b) >= 5 {
			return int64(int32(be.Uint32(b[1:5]))), 5, nil
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid integer length (%d)", len(b)),
			}
		}

	case ErlTypeSmallBig, ErlTypeLargeBig:
		var v *big.Int
		v, size, err = BigInt(b)

		if err == nil {
			ret = v.Int64()

			if v.Cmp(big.NewInt(ret)) != 0 {
				err = StructuralError{"integer too large"}
			}
		}

	default:
		err = SyntaxError{"not an integer"}
	}

	return
}

func UInt64(b []byte) (ret uint64, size int, err error) {
	var result int64
	result, size, err = Int64(b)
	ret = uint64(result)

	return
}

func String(b []byte) (ret string, size int, err error) {
	switch ErlType(b[0]) {
	case ErlTypeString:
		// $kLL…
		if len(b) >= 3 {
			size = 3 + int(be.Uint16(b[1:3]))

			if len(b) >= size {
				ret = string(b[3:size])
			} else {
				err = StructuralError{
					fmt.Sprintf("invalid string size (%d), len=%d", size, len(b)),
				}
			}
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid string length (%d)", len(b)),
			}
		}

	case ErlTypeBinary:
		var r []byte
		r, size, err = Binary(b)
		ret = string(r)

	case ErlTypeList:
		// $lLLLL…$j
		if len(b) > 5 {
			strLen := uint(be.Uint32(b[1:5]))
			size = 5
			b = b[size:]

			err = StructuralError{"not a string"}

			for i := uint(0); i < strLen; i++ {
				if len(b) <= 0 {
					err = StructuralError{"string ends abruptly"}
					return
				}

				switch ErlType(b[0]) {
				case ErlTypeSmallInteger, ErlTypeInteger, ErlTypeSmallBig, ErlTypeLargeBig:
					var char int64
					var charSize int
					var charErr error
					char, charSize, charErr = Int64(b)

					if charErr == nil {
						ret += string(char)
						size += charSize
						b = b[charSize:]
					} else {
						return
					}

				default:
					return
				}
			}

			if len(b) < 1 || ErlType(b[0]) != ErlTypeNil {
				err = StructuralError{"not a string (got improper list)"}
				return
			}

			size++
			err = nil
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid list length (%d)", len(b)),
			}
		}

	case ErlTypeNil:
		// $j
		size++

	default:
		err = SyntaxError{"not a string"}
	}

	return
}

func reverseBytes(b []byte) []byte {
	size := len(b)
	r := make([]byte, size)

	for i := 0; i < size; i++ {
		r[i] = b[size-i-1]
	}

	return r
}
