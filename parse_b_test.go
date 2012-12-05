package etf

import (
	"bytes"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

func Benchmark_parseAtom(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 64
	length := 64
	atoms := make([][]byte, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		s := bytes.Repeat([]byte{'a'}, length)
		writeAtom(w, Atom(string(bytes.Map(func(rune) rune { return rune(byte(rand.Int())) }, s))))
		atoms[i] = w.Bytes()
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := parseAtom(atoms[i%max])

		if err != nil {
			b.Fatal("failed to parse atom")
		}
	}
}

func Benchmark_parseBigInt(b *testing.B) {
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
		writeBigInt(w, new(big.Int).Sub(a, b))
		bigints[i] = w.Bytes()
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := parseBigInt(bigints[i%max])

		if err != nil {
			b.Fatal("failed to parse big int")
		}
	}
}

func Benchmark_parseFloat64(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 512
	floats := make([][]byte, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		writeFloat64(w, rand.ExpFloat64()-rand.ExpFloat64())
		floats[i] = w.Bytes()
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := parseFloat64(floats[i%max])

		if err != nil {
			b.Fatal("failed to parse float")
		}
	}
}

func Benchmark_parseString(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 64
	length := 64
	strings := make([][]byte, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		s := bytes.Repeat([]byte{'a'}, length)
		writeString(w, string(bytes.Map(func(rune) rune { return rune(byte(rand.Int())) }, s)))
		strings[i] = w.Bytes()
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := parseString(strings[i%max])

		if err != nil {
			b.Fatal("failed to parse string")
		}
	}
}
