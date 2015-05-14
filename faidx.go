package faidx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/biogo/biogo/io/seqio/fai"
)

type Faidx struct {
	rdr   io.ReadSeeker
	index fai.Index
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
	return &Faidx{rdr, idx}, nil
}

// Get a positions and returns the string sequence. Start and end are 1-based.
func (f *Faidx) Get(chrom string, start int, end int) (string, error) {
	idx, ok := f.index[chrom]
	if !ok {
		return "", fmt.Errorf("unknown sequence %s", chrom)
	}

	pstart, pend := idx.Position(start-1), idx.Position(end-1)+1
	f.rdr.Seek(pstart, 0)
	buf := make([]byte, pend-pstart)
	_, err := f.rdr.Read(buf)
	if err != nil {
		return "", err
	}
	buf = bytes.Replace(buf, []byte{'\n'}, []byte{}, -1)
	return string(buf), nil
}
