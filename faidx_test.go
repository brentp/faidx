package faidx_test

import (
	"testing"

	"github.com/brentp/faidx"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type FaidxTest struct{}

var _ = Suite(&FaidxTest{})

func (s *FaidxTest) TestNew(c *C) {
	fai, err := faidx.New("ce.fa")
	c.Assert(err, IsNil)
	c.Assert(fai, Not(IsNil))
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

func (s *FaidxTest) TestSeqs(c *C) {
	fai, err := faidx.New("ce.fa")
	c.Assert(err, IsNil)

	for _, test := range faiTests {
		seq, err := fai.Get(test.chrom, test.start, test.end)
		c.Assert(err, IsNil)
		c.Assert(seq, Equals, test.expected)

	}
}
