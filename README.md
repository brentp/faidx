faidx reader for golang using biogo's io.seqio.fai

```golang
f, err := faidx.New("some.fasta") 
check(err)

seq, err := f.Get("chr1", 1234, 4444)
```
