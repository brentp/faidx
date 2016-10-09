package faidx_test

import (
	"log"
	"math/rand"
	"testing"

	"github.com/brentp/faidx"
)

const file = "test.fa"

func benchmarkRandom(size int, b *testing.B, fai *faidx.Faidx) {
	idx := fai.Index["a"]
	mStart := idx.Length - size
	for n := 0; n < b.N; n++ {
		start := rand.Intn(mStart)
		fai.Get("a", start, start+size)
		fai.Get("g", 1, 100)
	}
}

func benchmarkSequential(size int, b *testing.B, fai *faidx.Faidx) {
	idx := fai.Index["a"]
	mStart := idx.Length - size
	step := 100
	for n := 0; n < b.N; n++ {
		for i := 0; i < mStart-step; i += step {
			fai.Get("a", i, i+size)
		}
	}
}

func getFai() *faidx.Faidx {
	fai, err := faidx.New(file)
	if err != nil {
		log.Fatal(err)
	}
	return fai
}

//func BenchmarkMmapRandom1(b *testing.B) { benchmarkRandom(1, b, getFai()) }
func BenchmarkMmapRandom10(b *testing.B) { benchmarkRandom(10, b, getFai()) }

//func BenchmarkMmapRandom100(b *testing.B) { benchmarkRandom(100, b, getFai()) }
//func BenchmarkMmapRandom500(b *testing.B)  { benchmarkRandom(500, b, getFai()) }
func BenchmarkMmapRandom1000(b *testing.B) { benchmarkRandom(1000, b, getFai()) }

//func BenchmarkMmapSequential1(b *testing.B) { benchmarkSequential(1, b, getFai()) }
func BenchmarkMmapSequential10(b *testing.B) { benchmarkSequential(10, b, getFai()) }

//func BenchmarkMmapSequential100(b *testing.B) { benchmarkSequential(100, b, getFai()) }
//func BenchmarkMmapSequential500(b *testing.B)  { benchmarkSequential(500, b, getFai()) }
func BenchmarkMmapSequential1000(b *testing.B) { benchmarkSequential(1000, b, getFai()) }
