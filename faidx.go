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

// Faidx is used to provide random access to the sequence data.
type Faidx struct {
	rdr   io.ReadSeeker
	Index fai.Index
	mmap  mmap.MMap
}

// ErrorNoFai is returned if the fasta doesn't have an associated .fai
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
		panic(fmt.Sprintf("fai: index [%d] out of range", p))
	}
	return r.Start + int64(p/r.BasesPerLine*r.BytesPerLine+p%r.BasesPerLine)
}

// Get takes a position and returns the string sequence. Start and end are 0-based.
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

// Stats hold sequenc information.
type Stats struct {
	// GC content fraction
	GC float64
	// CpG content fraction
	CpG float64
	// masked (lower-case fraction
	Masked float64
}

func min(a, b float64) float64 {
	if b < a {
		return b
	}
	return a
}

// Stats returns the proportion of GC's (GgCc), the CpG content (Cc follow by Gg)
// and the proportion of lower-case bases (masked).
// CpG will be 1.0 if the requested sequence is CGC and the base that follows is G
func (f *Faidx) Stats(chrom string, start int, end int) (Stats, error) {
	// copied from cnvkit.
	idx, ok := f.Index[chrom]
	if !ok {
		return Stats{}, fmt.Errorf("unknown sequence %s", chrom)
	}
	pstart := position(idx, start)
	pend := position(idx, end)
	oend := pend
	if pend < int64(len(f.mmap)) {
		oend++
	}

	var gcUp, gcLo, atUp, atLo, cpg int
	buf := f.mmap[pstart:oend]
	for i, v := range buf {
		// we added 1 to do the GC content...
		if i == len(buf)-1 {
			break
		}
		if v == 'G' || v == 'C' {
			if v == 'C' && (buf[i+1] == 'G' || buf[i+1] == 'g') {
				cpg++
			}
			gcUp++
		} else if v == 'A' || v == 'T' {
			atUp++
		} else if v == 'g' || v == 'c' {
			if v == 'c' && (buf[i+1] == 'G' || buf[i+1] == 'g') {
				cpg++
			}
			gcLo++
		} else if v == 'a' || v == 't' {
			atLo++
		}
	}
	tot := float64(gcUp + gcLo + atUp + atLo)
	if tot == 0.0 {
		return Stats{}, nil
	}
	return Stats{
		GC:     float64(gcLo+gcUp) / tot,
		Masked: float64(atLo+gcLo) / tot,
		CpG:    min(1.0, float64(2*cpg)/tot)}, nil
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
