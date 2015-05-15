package faidx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/biogo/biogo/io/seqio/fai"
	"github.com/edsrzf/mmap-go"
)

type Faidx struct {
	rdr   io.ReadSeeker
	Index fai.Index
	mmap  mmap.MMap
}

var ErrorNoFai = errors.New("no fai for fasta")

func notExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}

// New returns a faidx object from a fasta file that has an existing index.
func New(fasta string) (*Faidx, error) {
	err := notExists(fasta + ".fai")
	if err != nil {
		return nil, ErrorNoFai
	}
	fh, err := os.Open(fasta + ".fai")
	if err != nil {
		return nil, err
	}
	idx, err := fai.ReadFrom(fh)
	if err != nil {
		return nil, err
	}
	rdr, err := os.Open(fasta)
	if err != nil {
		return nil, err
	}

	smap, err := mmap.Map(rdr, mmap.RDONLY, 0)
	if err != nil {
		return nil, err
	}

	return &Faidx{rdr, idx, smap}, nil
}

func position(r fai.Record, p int) int64 {
	if p < 0 || r.Length < p {
		panic("fai: index out of range")
	}
	return r.Start + int64(p/r.BasesPerLine*r.BytesPerLine+p%r.BasesPerLine)
}

// At takes a position and returns the string sequence. Start and end are 0-based.
func (f *Faidx) Get(chrom string, start int, end int) (string, error) {
	idx, ok := f.Index[chrom]
	if !ok {
		return "", fmt.Errorf("unknown sequence %s", chrom)
	}

	pstart := position(idx, start)
	pend := position(idx, end)
	buf := f.mmap[pstart:pend]
	buf = bytes.Replace(buf, []byte{'\n'}, []byte{}, -1)
	return string(buf), nil
}

// At takes a single point and returns the single base.
func (f *Faidx) At(chrom string, pos int) (byte, error) {
	idx, ok := f.Index[chrom]
	if !ok {
		return '*', fmt.Errorf("unknown sequence %s", chrom)
	}

	ppos := position(idx, pos)
	return f.mmap[ppos], nil
}

// Close the associated Reader.
func (f *Faidx) Close() {
	f.rdr.(io.Closer).Close()
	f.mmap.Unmap()
}
