[![Build Status](https://travis-ci.org/brentp/faidx.svg?branch=master)](https://travis-ci.org/brentp/faidx)

faidx reader for golang using biogo's io.seqio.fai

```golang
f, err := faidx.New("some.fasta") 
check(err)

seq, err := f.Get("chr1", 1234, 4444)

st, err := f.Stats("chr1", 1234, 4444)

// fractions of GC content, CpG content and masked (lower-case)
st.GC, st.CpG, st.Masked
```
