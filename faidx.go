package faidx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
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
		panic(fmt.Sprintf("fai: index [%d] out of range in %s which has length: %d", p, r.Name, r.Length))
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

// FaPos allows the user to specify the position and internally, faidx will
// store information in it to speed GC calcs to adjacent regions. Useful for, when
// we sweep along the genome 1 base at a time, but we want to know the GC content for
// a window around each base.
type FaPos struct {
	Chrom string
	Start int
	End   int

	lastChrom string
	lastStart int
	lastEnd   int
	As        uint32
	Cs        uint32
	Gs        uint32
	Ts        uint32
}

// Duplicity returns a scaled entropy value of the counts of each base in p.
// Values approaching 1 are repetitive sequence values close to 0 have a more
// even distribution among the bases. This is likely to be called after `Q()`
// which populates the base-counts.
func (p *FaPos) Duplicity() float32 {
	n := float64(p.As + p.Cs + p.Gs + p.Ts)
	var s float64
	k := 0

	if p.As > 0 {
		s = float64(p.As) / n * math.Log(float64(p.As)/n)
		k++
	}
	if p.Cs > 0 {
		s += float64(p.Cs) / n * math.Log(float64(p.Cs)/n)
		k++
	}
	if p.Gs > 0 {
		s += float64(p.Gs) / n * math.Log(float64(p.Gs)/n)
		k++
	}
	if p.Ts > 0 {
		s += float64(p.Ts) / n * math.Log(float64(p.Ts)/n)
		k++
	}
	if k == 0 {
		return 0
	}
	if s == 0 {
		return 1.0
	}

	return float32(1 + s/math.Log(4))
}

// Q returns only the count of GCs it can do the calculation quickly for
// repeated calls marching to higher bases along the genome. It also
// updates the number of As, Cs, Ts, and Gs in FaPosition so the user
// can then calculate Entropy or use Duplicity above.
func (f *Faidx) Q(pos *FaPos) (uint32, error) {
	// we can't use any info from the cache
	idx, ok := f.Index[pos.Chrom]
	if !ok {
		return 0, fmt.Errorf("GC: unknown sequence %s", pos.Chrom)
	}

	if pos.lastStart > pos.Start || pos.Start >= pos.lastEnd || pos.lastEnd > pos.End || pos.Chrom != pos.lastChrom {
		pos.lastChrom = pos.Chrom
		pos.As, pos.Cs, pos.Gs, pos.Ts = 0, 0, 0, 0
		for _, b := range f.mmap[position(idx, pos.Start):position(idx, pos.End)] {
			switch b {
			case 'G', 'g':
				pos.Gs++
			case 'C', 'c':
				pos.Cs++
			case 'A', 'a':
				pos.As++
			case 'T', 't':
				pos.Ts++
			}
		}
	} else {
		/*
		 ls -------------- le
		       s----------------e
		*/
		for _, b := range f.mmap[position(idx, pos.lastStart):position(idx, pos.Start)] {
			switch b {
			case 'G', 'g':
				pos.Gs--
			case 'C', 'c':
				pos.Cs--
			case 'A', 'a':
				pos.As--
			case 'T', 't':
				pos.Ts--
			}
			/*
				if b == 'G' || b == 'g' {
					pos.Gs--
				} else if b == 'C' || b == 'c' {
					pos.Cs--
				} else if b == 'A' || b == 'a' {
					pos.As--
				} else if b == 'T' || b == 't' {
					pos.Ts--
				}
			*/
		}
		for _, b := range f.mmap[position(idx, pos.lastEnd):position(idx, pos.End)] {
			switch b {
			case 'G', 'g':
				pos.Gs++
			case 'C', 'c':
				pos.Cs++
			case 'A', 'a':
				pos.As++
			case 'T', 't':
				pos.Ts++
			}
		}

	}
	pos.lastStart = pos.Start
	pos.lastEnd = pos.End
	return pos.Gs + pos.Cs, nil
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
