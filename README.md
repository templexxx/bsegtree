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

## Performance

Tested on my WSL env:

```shell
➜  bsegtree git:(main) ✗ go test -v -run=^a -bench=.
goos: linux
goarch: amd64
pkg: github.com/templexxx/bsegtree
cpu: Intel(R) Core(TM) i5-8500 CPU @ 3.00GHz
BenchmarkBuildSmallTree
BenchmarkBuildSmallTree-6                 269674              3762 ns/op
BenchmarkBuildMidTree
BenchmarkBuildMidTree-6                     1405            821122 ns/op
BenchmarkQueryFullTree
BenchmarkQueryFullTree-6                  348902              3253 ns/op
BenchmarkQueryFullTreeSerial
BenchmarkQueryFullTreeSerial-6            365487              3239 ns/op
BenchmarkQueryPartTree
BenchmarkQueryPartTree/1_result
BenchmarkQueryPartTree/1_result-6       13721766                87.73 ns/op
BenchmarkQueryPartTree/4_result
BenchmarkQueryPartTree/4_result-6        5614016               210.8 ns/op
BenchmarkQueryPartTree/16_result
BenchmarkQueryPartTree/16_result-6       2201541               543.4 ns/op
BenchmarkQueryPartTree/64_result
BenchmarkQueryPartTree/64_result-6        942684              1248 ns/op
BenchmarkQueryPartTree/256_result
BenchmarkQueryPartTree/256_result-6       644641              1668 ns/op
BenchmarkQueryPartTree/1024_result
BenchmarkQueryPartTree/1024_result-6      309175              3247 ns/op
BenchmarkQueryPartTreeSerial
BenchmarkQueryPartTreeSerial/1_result
BenchmarkQueryPartTreeSerial/1_result-6                  1426617               838.4 ns/op
BenchmarkQueryPartTreeSerial/4_result
BenchmarkQueryPartTreeSerial/4_result-6                  1434604               836.1 ns/op
BenchmarkQueryPartTreeSerial/16_result
BenchmarkQueryPartTreeSerial/16_result-6                 1389748               863.0 ns/op
BenchmarkQueryPartTreeSerial/64_result
BenchmarkQueryPartTreeSerial/64_result-6                 1208978               986.5 ns/op
BenchmarkQueryPartTreeSerial/256_result
BenchmarkQueryPartTreeSerial/256_result-6                 754640              1434 ns/op
BenchmarkQueryPartTreeSerial/1024_result
BenchmarkQueryPartTreeSerial/1024_result-6                317222              3175 ns/op
BenchmarkQueryPoint
BenchmarkQueryPoint-6                                   11519022                99.12 ns/op
BenchmarkQueryPointSerial
BenchmarkQueryPointSerial-6                              1428042               837.5 ns/op
BenchmarkQueryPointSerialCapacity
BenchmarkQueryPointSerialCapacity/4
BenchmarkQueryPointSerialCapacity/4-6                   44112619                26.11 ns/op
BenchmarkQueryPointSerialCapacity/16
BenchmarkQueryPointSerialCapacity/16-6                  32856278                34.68 ns/op
BenchmarkQueryPointSerialCapacity/64
BenchmarkQueryPointSerialCapacity/64-6                  14142586                83.04 ns/op
BenchmarkQueryPointSerialCapacity/256
BenchmarkQueryPointSerialCapacity/256-6                  5421559               221.6 ns/op
BenchmarkQueryPointSerialCapacity/1024
BenchmarkQueryPointSerialCapacity/1024-6                 1551112               772.8 ns/op
BenchmarkQueryPointCapacity
BenchmarkQueryPointCapacity/4
BenchmarkQueryPointCapacity/4-6                         38701812                30.69 ns/op
BenchmarkQueryPointCapacity/16
BenchmarkQueryPointCapacity/16-6                        27751310                42.75 ns/op
BenchmarkQueryPointCapacity/64
BenchmarkQueryPointCapacity/64-6                        14492997                83.11 ns/op
BenchmarkQueryPointCapacity/256
BenchmarkQueryPointCapacity/256-6                       12500755                94.16 ns/op
BenchmarkQueryPointCapacity/1024
BenchmarkQueryPointCapacity/1024-6                      11824395               102.3 ns/op
PASS
ok      github.com/templexxx/bsegtree   41.750s
```

Fast enough for not big intervals (<= 1024) query, have been met my needs.