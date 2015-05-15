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

/*
func (s *FaidxTest) TestMAt(c *C) {
	for _, test := range faiTests {
		seq, err := s.fai.MAt(test.chrom, test.start-1, test.end)
		c.Assert(err, IsNil)
		c.Assert(seq, Equals, test.expected)

	}
}*/
