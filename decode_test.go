package etf

import (
	"github.com/ftrvxmtrx/testingo"
	"math/big"
	"testing"
)

func Test_Decode_Array(t0 *testing.T) {
	t := testingo.T(t0)
	var data [3]byte

	size, err := Decode([]byte{131, 109, 0, 0, 0, 3, 1, 2, 3}, &data)
	t.AssertEq(nil, err)
	t.AssertEq(uint(9), size)
	t.AssertEq(byte(1), data[0])
	t.AssertEq(byte(2), data[1])
	t.AssertEq(byte(3), data[2])

	size, err = Decode([]byte{131, 107, 0, 3, 1, 2, 3}, &data)
	t.AssertEq(nil, err)
	t.AssertEq(uint(7), size)
	t.AssertEq(byte(1), data[0])
	t.AssertEq(byte(2), data[1])
	t.AssertEq(byte(3), data[2])
}

func Test_Decode_BigInt(t0 *testing.T) {
	t := testingo.T(t0)
	var bigint *big.Int

	size, err := Decode([]byte{131, 110, 15, 0, 0, 0, 0, 0, 16,
		159, 75, 179, 21, 7, 201, 123, 206, 151, 192,
	}, &bigint)
	t.AssertEq(nil, err)
	t.AssertEq(uint(19), size)
}

func Test_Decode_Binary(t0 *testing.T) {
	t := testingo.T(t0)
	var data []byte

	size, err := Decode([]byte{131, 109, 0, 0, 0, 3, 1, 2, 3}, &data)
	t.AssertEq(nil, err)
	t.AssertEq(uint(9), size)
	t.AssertEq(byte(1), data[0])
	t.AssertEq(byte(2), data[1])
	t.AssertEq(byte(3), data[2])
}

func Test_Decode(t0 *testing.T) {
	t := testingo.T(t0)
	var s string
	size, err := Decode([]byte{131, 107, 0, 3, 49, 50, 51}, &s)
	t.AssertEq(nil, err)
	t.AssertEq(uint(7), size)
	t.AssertEq("123", s)

	type testStruct struct {
		Atom
		X uint8
		S string
	}

	var ts testStruct

	size, err = Decode([]byte{
		131, 104, 3, 100, 0, 4, 98, 108, 97, 104, 97, 4, 108, 0, 0, 0, 4, 98,
		0, 0, 4, 68, 98, 0, 0, 4, 75, 98, 0, 0, 4, 50, 98, 0, 0, 4, 48, 106,
	}, &ts)
	t.AssertEq(nil, err)
	t.AssertEq(uint(38), size)
	t.AssertEq(uint8(4), ts.X)
	t.AssertEq("фыва", ts.S)

	size, err = Decode([]byte{
		131, 104, 3, 99, 50, 46, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57,
		57, 57, 56, 56, 56, 57, 56, 101, 45, 48, 49, 0, 0, 0, 0, 0, 97, 4, 108, 0, 0,
		0, 4, 98, 0, 0, 4, 68, 98, 0, 0, 4, 75, 98, 0, 0, 4, 50, 98, 0, 0, 4, 48, 106,
	}, &ts)
	t.AssertNotEq(nil, err)

	type testStruct2 struct {
		T testStruct
		Y int
	}

	var ts2 testStruct2

	size, err = Decode([]byte{
		131, 104, 2, 104, 3, 100, 0, 4, 98, 108, 97, 104, 97, 4, 108, 0, 0, 0, 4, 98,
		0, 0, 4, 68, 98, 0, 0, 4, 75, 98, 0, 0, 4, 50, 98, 0, 0, 4, 48, 106, 98, 0, 0, 2, 154,
	}, &ts2)
	t.AssertEq(nil, err)
	t.AssertEq(uint(45), size)
	t.AssertEq(uint8(4), ts2.T.X)
	t.AssertEq("фыва", ts2.T.S)
	t.AssertEq(666, ts2.Y)
}
