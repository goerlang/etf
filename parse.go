package etf

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

var be = binary.BigEndian

const (
	errPrefix = "Erlang external format "
)

type StructuralError struct {
	Msg string
}

type SyntaxError struct {
	Msg string
}

func (err StructuralError) Error() string {
	return errPrefix + "structural error: " + err.Msg
}

func (err SyntaxError) Error() string {
	return errPrefix + "syntax error: " + err.Msg
}

func parseAtom(b []byte) (ret Atom, size uint, err error) {
	switch erlType(b[0]) {
	case erlAtom:
		// $dLL…
		if len(b) >= 3 {
			size = 3 + uint(be.Uint16(b[1:3]))

			if uint(len(b)) >= size {
				ret = Atom(b[3:size])
			} else {
				err = StructuralError{"wrong atom size"}
			}
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid atom length (%d)", len(b)),
			}
		}

	case erlSmallAtom:
		// $sL…
		if len(b) >= 2 {
			size = 2 + uint(b[1])

			if uint(len(b)) >= size {
				ret = Atom(b[2:size])
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

func parseBigInt(b []byte) (ret *big.Int, size uint, err error) {
	switch erlType(b[0]) {
	case erlSmallBig:
		// $nAS…
		if len(b) >= 3 && len(b)-3 >= int(b[1]) {
			size = 3 + uint(b[1])
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

	case erlLargeBig:
		// $oAAAAS…
		if len(b) >= 6 {
			length := uint(be.Uint32(b[1:5]))
			size = 6 + length

			if uint(len(b)) >= size && uint(int32(length)+6) == length+6 {
				b2 := reverseBytes(b[6 : 6+int(length)])
				ret = new(big.Int).SetBytes(b2)

				if b[5] != 0 {
					ret = ret.Neg(ret)
				}
			} else {
				err = StructuralError{"invalid large bigInt"}
			}
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid large big integer length (%d)", len(b)),
			}
		}

	default:
		err = SyntaxError{"not a big integer"}
	}

	return
}

func parseBinary(b []byte) (ret []byte, size uint, err error) {
	var s int

	switch t := erlType(b[0]); t {
	case erlBinary:
		// $mLLLL…
		s = 5

	case erlString:
		// $kLL…
		s = 3

	default:
		err = SyntaxError{"not a binary"}
	}

	if err == nil {
		if len(b) >= s {
			if s == 5 {
				size = uint(s) + uint(be.Uint32(b[1:s]))
			} else {
				size = uint(s) + uint(be.Uint16(b[1:s]))
			}

			if uint(len(b)) >= size {
				ret = b[s:size]
			} else {
				err = StructuralError{
					fmt.Sprintf("invalid binary size (%d), len=%d", size, len(b)),
				}
			}
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid binary length (%d)", len(b)),
			}
		}
	}

	return
}

func parseBool(b []byte) (ret bool, size uint, err error) {
	var v Atom

	v, size, err = parseAtom(b)

	if err == nil {
		switch v {
		case Atom("true"):
			ret = true
			return

		case Atom("false"):
			ret = false
			return
		}

		err = SyntaxError{"not a boolean"}
	}

	return
}

func parseFloat64(b []byte) (ret float64, size uint, err error) {
	switch erlType(b[0]) {
	case erlFloat:
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

	case erlNewFloat:
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

func parseInt64(b []byte) (ret int64, size uint, err error) {
	switch erlType(b[0]) {
	case erlSmallInteger:
		// $aI
		if len(b) >= 2 {
			return int64(b[1]), 2, nil
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid small integer length (%d)", len(b)),
			}
		}

	case erlInteger:
		// $bIIII
		if len(b) >= 5 {
			return int64(int32(be.Uint32(b[1:5]))), 5, nil
		} else {
			err = StructuralError{
				fmt.Sprintf("invalid integer length (%d)", len(b)),
			}
		}

	case erlSmallBig, erlLargeBig:
		var v *big.Int
		v, size, err = parseBigInt(b)

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

func parseUint64(b []byte) (ret uint64, size uint, err error) {
	var result int64
	result, size, err = parseInt64(b)
	ret = uint64(result)

	return
}

func parseString(b []byte) (ret string, size uint, err error) {
	switch erlType(b[0]) {
	case erlString:
		// $kLL…
		if len(b) >= 3 {
			size = 3 + uint(be.Uint16(b[1:3]))

			if uint(len(b)) >= size {
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

	case erlBinary:
		var r []byte
		r, size, err = parseBinary(b)
		ret = string(r)

	case erlList:
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

				switch erlType(b[0]) {
				case erlSmallInteger, erlInteger, erlSmallBig, erlLargeBig:
					var char int64
					var charSize uint
					var charErr error
					char, charSize, charErr = parseInt64(b)

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

			if len(b) < 1 || erlType(b[0]) != erlNil {
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

	case erlNil:
		// $j
		size++

	default:
		err = SyntaxError{"not a string"}
	}

	return
}
