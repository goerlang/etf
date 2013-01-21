package read

import (
	"bytes"
	"encoding/binary"
	t "github.com/goerlang/etf/types"
	"io"
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
	atoms := make([]*bytes.Buffer, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		s := bytes.Repeat([]byte{'a'}, length)
		b := bytes.Map(randRune, s)
		w.Write([]byte{t.EttSmallAtom, byte(length)})
		w.Write(b)
		atoms[i] = w
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := atoms[i%max]
		_, err := Atom(in)

		if err != io.EOF && err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBigInt(b *testing.B) {
	b.StopTimer()

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	uint64Max := big.NewInt(math.MaxInt64)
	top := new(big.Int).Mul(uint64Max, uint64Max)
	max := 512
	bigints := make([]*bytes.Buffer, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		a := new(big.Int).Rand(rand, top)
		b := new(big.Int).Rand(rand, top)
		v := new(big.Int).Sub(a, b)
		sign := 0
		if v.Sign() < 0 {
			sign = 1
		}
		bytes := reverse(new(big.Int).Abs(a).Bytes())
		w.Write([]byte{t.EttSmallBig, byte(len(bytes)), byte(sign)})
		w.Write(bytes)
		bigints[i] = w
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := bigints[i%max]
		_, err := BigInt(in)

		if err != io.EOF && err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBinary(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 64
	length := 64
	binaries := make([]*bytes.Buffer, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		s := bytes.Repeat([]byte{'a'}, length)
		b := bytes.Map(func(rune) rune { return rune(byte(rand.Int())) }, s)
		w.Write([]byte{t.EttBinary})
		binary.Write(w, binary.BigEndian, uint32(len(b)))
		w.Write(b)
		binaries[i] = w
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := binaries[i%max]
		_, err := Binary(in)

		if err != io.EOF && err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFloat(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 512
	floats := make([]*bytes.Buffer, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		v := rand.ExpFloat64() - rand.ExpFloat64()
		w.Write([]byte{t.EttNewFloat})
		binary.Write(w, binary.BigEndian, math.Float64bits(v))
		floats[i] = w
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := floats[i%max]
		_, err := Float(in)

		if err != io.EOF && err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPid(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 64
	length := 16
	pids := make([]*bytes.Buffer, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		s := bytes.Repeat([]byte{'a'}, length)
		b := bytes.Map(randRune, s)
		b[6] = '@'
		w.Write([]byte{t.EttPid, t.EttSmallAtom, byte(length)})
		w.Write(b)
		w.Write([]byte{0, 0, 0, uint8(rand.Int())})
		w.Write([]byte{0, 0, 0, uint8(rand.Int())})
		w.Write([]byte{uint8(rand.Int())})
		pids[i] = w
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := pids[i%max]
		_, err := Pid(in)

		if err != io.EOF && err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkString(b *testing.B) {
	b.StopTimer()

	rand.Seed(time.Now().UnixNano())
	max := 64
	length := 64
	strings := make([]*bytes.Buffer, max)

	for i := 0; i < max; i++ {
		w := new(bytes.Buffer)
		s := bytes.Repeat([]byte{'a'}, length)
		b := bytes.Map(randRune, s)
		w.Write([]byte{t.EttString})
		binary.Write(w, binary.BigEndian, uint16(len(b)))
		w.Write(b)
		strings[i] = w
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		in := strings[i%max]
		_, err := String(in)

		if err != io.EOF && err != nil {
			b.Fatal(err)
		}
	}
}

func randRune(_ rune) rune {
	return rune('0' + byte(rand.Intn('z'-'0'+1)))
}
