package faidx_test

import (
	"testing"

	"github.com/brentp/faidx"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type FaidxTest struct {
	fai *faidx.Faidx
}

var _ = Suite(&FaidxTest{})

func (s *FaidxTest) SetUpTest(c *C) {
	fai, err := faidx.New("test.fa")
	c.Assert(err, IsNil)
	c.Assert(fai, Not(IsNil))
	s.fai = fai
}

var faiTests = []struct {
	chrom    string
	start    int
	end      int
	expected string
}{
	{"a", 100, 201, "TAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCC"},
	{"a", 141, 201, "CTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCC"},
	{"a", 142, 201, "TAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCC"},
	{"a", 142, 200, "TAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGCCTAAGC"},
	{"d", 1, 10, "CCTAAGCCTA"},
	{"f", 4996, 5000, "GTCTC"},
	{"g", 4996, 5000, "TTTGG"},
}

func (s *FaidxTest) TestGet(c *C) {
	for _, test := range faiTests {
		seq, err := s.fai.Get(test.chrom, test.start-1, test.end)
		c.Assert(err, IsNil)
		c.Assert(seq, Equals, test.expected)

	}
}

func (s *FaidxTest) TestStats(c *C) {
	st, err := s.fai.Stats("f", 4995, 5000)
	c.Assert(err, IsNil)
	c.Assert(st.Masked, Equals, 0.0)
	c.Assert(st.CpG, Equals, 0.0)
	c.Assert(st.GC, Equals, 0.6)

	seq, err := s.fai.Get("a", 103-1, 110)
	c.Assert(seq, Equals, "GCCTAAGC")
	c.Assert(err, Equals, nil)

	st, err = s.fai.Stats("a", 103-1, 110)
	c.Assert(err, IsNil)
	c.Assert(st.Masked, Equals, 0.0)
	c.Assert(st.CpG, Equals, 0.0)
	c.Assert(st.GC, Equals, float64(5)/float64(8))

	st, err = s.fai.Stats("g", 4996-1, 5000)
	c.Assert(err, IsNil)
	c.Assert(st.Masked, Equals, 0.0)
	c.Assert(st.CpG, Equals, 0.0)
	c.Assert(st.GC, Equals, float64(2)/float64(5))

	seq, err = s.fai.Get("k", 0, 9)
	c.Assert(seq, Equals, "CGCGCGCGA")

	st, err = s.fai.Stats("k", 0, 9)
	c.Assert(st.CpG, Equals, float64(2*4)/float64(9))
	c.Assert(err, IsNil)
}

func (s *FaidxTest) TestCpG(c *C) {
	seq, err := s.fai.Get("k", 0, 2)
	c.Assert(seq, Equals, "CG")
	c.Assert(err, IsNil)
	st, err := s.fai.Stats("k", 0, 2)
	c.Assert(err, IsNil)
	c.Assert(st.CpG, Equals, 1.0)

	st, err = s.fai.Stats("k", 6, 9)
	seq, err = s.fai.Get("k", 6, 9)
	c.Assert(st.CpG, Equals, 2.0/3.0)

	st, err = s.fai.Stats("k", 0, 1)
	seq, err = s.fai.Get("k", 0, 1)
	c.Assert(err, IsNil)
	c.Assert(st.CpG, Equals, 1.0)
}
