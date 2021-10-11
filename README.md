# bsegtree
Segment tree for bytes in Go

Based on [Thomas Oberndörfer's int range segment tree](https://github.com/toberndo/go-stree) with fixing/optimization/modification for bytes ranges.

1. For build once, query many models
2. Not design for big data set 
(better for <= 1024 intervals, otherwise the cost will be quite high when query a big range. Bench it before using. 
For small count intervals. e.g., 1024, the point query will be ~100ns if only one interval id will be returned)
3. Not design for interval has long bytes as range
4. Using pseudo-codes in [Computational Geometry: Algorithms and Applications, INSERTSEGMENTTREE](http://www.cs.uu.nl/geobook/pseudo.pdf) &
codes in [another segment tree implementation](https://github.com/seppestas/go-segtree) to fix
wrong query result for some corner cases. See `func TestMissingResult(t *testing.T)` in [bstree_test.go](bstree_test.go)

## Details of Implementation

1. Using uint64 as abbreviated key for speeding up query & push, which means long keys with long common prefix won't work with this lib.
2. Build is slow, offline building is preferred in production environment.
3. Invoker has responsibility to map the id and target, query will only return the id. ID is started from 0, each push will plus 1.

## Pefromance

```shell
➜  bsegtree git:(main) ✗ go test -v -run=^a -bench=.
goos: darwin
goarch: amd64
pkg: github.com/templexxx/bsegtree
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkBuildSmallTree
BenchmarkBuildSmallTree-8        	  277869	      3983 ns/op
BenchmarkBuildMidTree
BenchmarkBuildMidTree-8          	    1310	    900920 ns/op
BenchmarkQueryFullTree
BenchmarkQueryFullTree-8         	    5900	    191030 ns/op
BenchmarkQueryPartTree
BenchmarkQueryPartTree-8         	  794811	      1453 ns/op
BenchmarkQueryPoint
BenchmarkQueryPoint-8            	 9096208	       126.7 ns/op
BenchmarkQueryFullTreeSerial
BenchmarkQueryFullTreeSerial-8   	  167599	      6600 ns/op
BenchmarkQueryPartTreeSerial
BenchmarkQueryPartTreeSerial-8   	  829184	      1453 ns/op
BenchmarkQueryPointSerial
BenchmarkQueryPointSerial-8      	 1000000	      1125 ns/op
PASS
ok  	github.com/templexxx/bsegtree	10.825s
```

For point query (1024 intervals) for 1 interval result is good enough.

But for more than 8 interval result (1024 intervals) the serial begins to be better.