package faidx_test

import (
	"log"
	"math/rand"
	"testing"

	"github.com/brentp/faidx"
)

const file = "test.fa"

func benchmarkRandom(size int, b *testing.B, fai *faidx.Faidx, mmap bool) {
	idx := fai.Index["a"]
	mStart := idx.Length - size
	if mmap {
		for n := 0; n < b.N; n++ {
			start := rand.Intn(mStart)
			fai.MAt("a", start, start+size)
			fai.MAt("g", 1, 100)
		}
	} else {
		for n := 0; n < b.N; n++ {
			start := rand.Intn(mStart)
			fai.At("a", start, start+size)
			fai.At("g", 1, 100)
		}
	}
}

func benchmarkSequential(size int, b *testing.B, fai *faidx.Faidx, mmap bool) {
	idx := fai.Index["a"]
	mStart := idx.Length - size
	step := 100
	if mmap {
		for n := 0; n < b.N; n++ {
			for i := 0; i < mStart-step; i += step {
				fai.MAt("a", i, i+size)
			}
		}
	} else {
		for n := 0; n < b.N; n++ {
			for i := 0; i < mStart-step; i += step {
				fai.At("a", i, i+size)
			}
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

func BenchmarkSeekRandom1(b *testing.B)    { benchmarkRandom(1, b, getFai(), false) }
func BenchmarkSeekRandom10(b *testing.B)   { benchmarkRandom(10, b, getFai(), false) }
func BenchmarkSeekRandom100(b *testing.B)  { benchmarkRandom(100, b, getFai(), false) }
func BenchmarkSeekRandom500(b *testing.B)  { benchmarkRandom(500, b, getFai(), false) }
func BenchmarkSeekRandom1000(b *testing.B) { benchmarkRandom(1000, b, getFai(), false) }

func BenchmarkMmapRandom1(b *testing.B)    { benchmarkRandom(1, b, getFai(), true) }
func BenchmarkMmapRandom10(b *testing.B)   { benchmarkRandom(10, b, getFai(), true) }
func BenchmarkMmapRandom100(b *testing.B)  { benchmarkRandom(100, b, getFai(), true) }
func BenchmarkMmapRandom500(b *testing.B)  { benchmarkRandom(500, b, getFai(), true) }
func BenchmarkMmapRandom1000(b *testing.B) { benchmarkRandom(1000, b, getFai(), true) }

func BenchmarkSeekSequential1(b *testing.B)    { benchmarkSequential(1, b, getFai(), false) }
func BenchmarkSeekSequential10(b *testing.B)   { benchmarkSequential(10, b, getFai(), false) }
func BenchmarkSeekSequential100(b *testing.B)  { benchmarkSequential(100, b, getFai(), false) }
func BenchmarkSeekSequential500(b *testing.B)  { benchmarkSequential(500, b, getFai(), false) }
func BenchmarkSeekSequential1000(b *testing.B) { benchmarkSequential(1000, b, getFai(), false) }

func BenchmarkMmapSequential1(b *testing.B)    { benchmarkSequential(1, b, getFai(), true) }
func BenchmarkMmapSequential10(b *testing.B)   { benchmarkSequential(10, b, getFai(), true) }
func BenchmarkMmapSequential100(b *testing.B)  { benchmarkSequential(100, b, getFai(), true) }
func BenchmarkMmapSequential500(b *testing.B)  { benchmarkSequential(500, b, getFai(), true) }
func BenchmarkMmapSequential1000(b *testing.B) { benchmarkSequential(1000, b, getFai(), true) }
