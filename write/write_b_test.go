package write

import (
	"bytes"
	. "github.com/goerlang/etf/types"
	. "io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkAtom(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 64
	length := 64
	atoms := make([]ErlAtom, max)

	for i := 0; i < max; i++ {
		atoms[i] = ErlAtom(bytes.Repeat([]byte{byte('A' + i)}, length))
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := atoms[i%max]
		if err := Atom(Discard, in); err != nil {
			b.Fatal(in, err)
		}
	}
}

func BenchmarkBigInt(b *testing.B) {
	b.StopTimer()

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	uint64Max := big.NewInt(math.MaxInt64)
	top := new(big.Int).Mul(uint64Max, uint64Max)
	max := 512
	bigints := make([]*big.Int, max)

	for i := 0; i < max; i++ {
		a := new(big.Int).Rand(rand, top)
		b := new(big.Int).Rand(rand, top)
		bigints[i] = new(big.Int).Sub(a, b)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := bigints[i%max]
		if err := BigInt(Discard, in); err != nil {
			b.Fatal(in, err)
		}
	}
}

func BenchmarkBinary(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 64
	length := 64
	binaries := make([][]byte, max)

	for i := 0; i < max; i++ {
		s := bytes.Repeat([]byte{'a'}, length)
		binaries[i] = bytes.Map(
			func(rune) rune { return rune(byte(rand.Int())) },
			s,
		)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := binaries[i%max]
		if err := Binary(Discard, in); err != nil {
			b.Fatal(in, err)
		}
	}
}

func BenchmarkFloat64(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 512
	floats := make([]float64, max)

	for i := 0; i < max; i++ {
		floats[i] = rand.ExpFloat64() - rand.ExpFloat64()
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := floats[i%max]
		if err := Float64(Discard, in); err != nil {
			b.Fatal(in, err)
		}
	}
}

func BenchmarkInt64(b *testing.B) {
	b.StopTimer()

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	max := 512
	ints := make([]int64, max)

	for i := 0; i < max; i++ {
		ints[i] = rand.Int63() - rand.Int63()
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := ints[i%max]
		if err := Int64(Discard, in); err != nil {
			b.Fatal(in, err)
		}
	}
}

func BenchmarkString(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 64
	length := 64
	strings := make([]string, max)

	for i := 0; i < max; i++ {
		s := bytes.Repeat([]byte{'a'}, length)
		strings[i] = string(bytes.Map(
			func(rune) rune { return rune(byte(rand.Int())) },
			s,
		))
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := strings[i%max]
		if err := String(Discard, in); err != nil {
			b.Fatal(in, err)
		}
	}
}
