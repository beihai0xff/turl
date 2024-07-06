
# 性能测试

对 API 的性能测试，主要是为了验证 API 的性能是否满足需求，以及在不同的并发情况下，API 的性能表现。

## 测试资源

服务器：Apple MacBook Pro14 M1 Pro 2021 16G 512G SSD

## 测试方式

获取全球访问量前 10k 的域名，每个域名添加 10 个 API 后缀，共计 100k 条无重复数据。使用 100 个协程并发请求，统计写入耗时。

```shell
Benchmark_Create
    api_benchmark_test.go:93: send requests: 4550
    api_benchmark_test.go:93: send requests: 10040
    api_benchmark_test.go:93: send requests: 15422
    api_benchmark_test.go:93: send requests: 20977
    api_benchmark_test.go:93: send requests: 26532
    api_benchmark_test.go:93: send requests: 32452
    api_benchmark_test.go:93: send requests: 38095
    api_benchmark_test.go:93: send requests: 42921
    api_benchmark_test.go:93: send requests: 47579
    api_benchmark_test.go:93: send requests: 52629
    api_benchmark_test.go:93: send requests: 58464
    api_benchmark_test.go:93: send requests: 63961
    api_benchmark_test.go:93: send requests: 69651
    api_benchmark_test.go:93: send requests: 75325
    api_benchmark_test.go:93: send requests: 80952
    api_benchmark_test.go:93: send requests: 87016
    api_benchmark_test.go:93: send requests: 92535
    api_benchmark_test.go:93: send requests: 98344
    api_benchmark_test.go:111: success requests:  100000 costs 18.365219458s
```
综合来看，API 的性能表现良好，能够达到 5000+ QPS 的写入速度。