package parse

import (
	"bytes"
	"encoding/binary"
	. "github.com/goerlang/etf/types"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

func Benchmark_Atom(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 64
	length := 64
	atoms := make([][]byte, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		w.Write([]byte{ErlTypeSmallAtom, byte(length)})
		w.Write(bytes.Repeat([]byte{byte('A' + i)}, length))
		atoms[i] = w.Bytes()
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := atoms[i%max]
		_, _, err := Atom(in)

		if err != nil {
			b.Fatal("failed to parse atom %#v", in)
		}
	}
}

func Benchmark_BigInt(b *testing.B) {
	b.StopTimer()

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	uint64Max := big.NewInt(math.MaxInt64)
	top := new(big.Int).Mul(uint64Max, uint64Max)
	max := 512
	bigints := make([][]byte, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		a := new(big.Int).Rand(rand, top)
		b := new(big.Int).Rand(rand, top)
		v := new(big.Int).Sub(a, b)
		sign := 0
		if v.Sign() < 0 {
			sign = 1
		}
		bytes := reverseBytes(new(big.Int).Abs(a).Bytes())
		w.Write([]byte{ErlTypeSmallBig, byte(len(bytes)), byte(sign)})
		w.Write(bytes)
		bigints[i] = w.Bytes()
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := bigints[i%max]
		_, _, err := BigInt(in)

		if err != nil {
			b.Fatal("failed to parse big int %#v", in)
		}
	}
}

func Benchmark_Float64(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 512
	floats := make([][]byte, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		v := rand.ExpFloat64() - rand.ExpFloat64()
		w.Write([]byte{ErlTypeNewFloat})
		binary.Write(w, binary.BigEndian, math.Float64bits(v))
		floats[i] = w.Bytes()
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := floats[i%max]
		_, _, err := Float64(in)

		if err != nil {
			b.Fatal("failed to parse float %#v", in)
		}
	}
}

func Benchmark_String(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 64
	length := 64
	strings := make([][]byte, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		s := bytes.Repeat([]byte{'a'}, length)
		b := bytes.Map(func(rune) rune { return rune(byte(rand.Int())) }, s)
		w.Write([]byte{ErlTypeString})
		binary.Write(w, binary.BigEndian, uint16(len(b)))
		w.Write(b)
		strings[i] = w.Bytes()
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := strings[i%max]
		_, _, err := String(in)

		if err != nil {
			b.Fatal("failed to parse string %#v", in)
		}
	}
}
