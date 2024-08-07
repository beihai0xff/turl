[![codecov](https://codecov.io/gh/beihai0xff/turl/graph/badge.svg?token=DPVOTT6MIU)](https://codecov.io/gh/beihai0xff/turl)
[![GitHub Action](https://github.com/beihai0xff/turl/actions/workflows/ci.yml/badge.svg)](https://github.com/beihai0xff/turl/actions/)
[![Go Report Card](https://goreportcard.com/badge/github.com/beihai0xff/turl)](https://goreportcard.com/report/github.com/beihai0xff/turl)

# turl
Tiny-URL 短链接服务

在社交媒体、用户增长、广告投放等场景中，经常会遇到长链接转短链接的需求，提高用户点击率更高，同时能规避原始链接中一些关键词、域名屏蔽等。常见微博、微信等社交软件中，比如微博限制字数为140，如果包含的链接过长，会占用很多字数，所以需要将长链接转换为短链接，以节省字数。

短链接除了具有美观清爽的特性外，利用短链每次跳转都需要经过后端的特性，可以在跳转过程中做异步埋点，用于效果数据统计，常见的应用场景如下：

* 注册、收藏、加购、下单、支付效果统计；
* 用户分享效果追踪；
* 减少字符占用。

# Features

## 开发进度
- [x] 分布式 ID 生成器：基于 TDDL 生成唯一的 ID；
- [x] 分布式缓存：支持 Redis 缓存；
- [x] 本地缓存：支持 bigcache 本地缓存；
- [x] 数据库：支持 MySQL 数据库；
- [x] URL 302 重定向；
- [x] URL 编码：支持 Base58 编码；
- [x] 限流器：支持 Redis 与单机令牌桶限流器；
- [x] 读写分离：只读/只写/读写模式运行；
- [x] 幂等：同一 URL 多次生成，需要保证生成的短链接是唯一的；
- [ ] 过期时间：支持短链接过期时间；
- [ ] 可观测：API 访问数据数据、服务监控；

# 快速体验

## 本地运行

确保本地已经安装了 Docker 与 Docker Compose，然后执行以下命令：
```shell
make deploy
```

终端输出 `turl service containers start successfully` 后，说明服务已经启动成功。
该模式会部署 MySQL 与 Redis 服务，作为本地存储与缓存服务器。同时会启动两个服务节点，一个用于读写操作，另一个用于只读操作。
- 读写服务：[http://localhost:8080](http://localhost:8080)，用于生成短链接、更新远程缓存、更新数据库等；
- 只读服务：[http://localhost:80](http://localhost:80)，只用于访问短链接，不支持生成短链接，生产环境中可以部署多个只读服务节点，用于分流读取请求。
- swagger：访问 [http://localhost:8080/v1/management/swagger/index.html#/](http://localhost:8080/v1/management/swagger/index.html#/) swagger 页面，
## API 接口

### 生成短链接

```shell
curl -X POST http://localhost:8080/v1/management/shorten -H 'Content-Type: application/json' -d '{"long_url": "https://google.com"}'
```
返回结果：
```json
{"short_url":"http://localhost/24rgcX","long_url":"https://google.com","created_at":"2024-07-08T15:06:26.434Z","deleted_at":null,"error":""}
```

### 访问短链接

访问短链接 `http://localhost/24rgcX`，将会被重定向到原始的长链接 `https://google.com`。

```shell
curl -L http://localhost/24rgcX
```

### 获取长链接信息

```shell
curl -X GET http://localhost:8080/v1/management/shorten\?long_url\=https://google.com
```

返回结果：
```json
{"short_url":"http://localhost/24rgcX","long_url":"https://google.com","created_at":"2024-07-08T15:06:26.434Z","deleted_at":null,"error":""}
```


# 短链接服务系统设计

## 功能需求
* 短链接生成：给定一个长链接，能够生成一个唯一的短链接，即使多次生成同一个长链接，也能保证生成的短链接是唯一的。
* 短链接重定向：通过短链接能够访问到原始的长链接，通过 302 临时重定向的方式，将用户重定向到原始的长链接，临时重定向的方式可以保证搜索引擎不会抓取短链接，而是抓取原始的长链接，并且便于统计短链接的访问次数。
* 访问限流：对短链接的访问可以设置限流，限制每个短链接单位时间内的的访问次数。
* 过期时间：短链接可以设置过期时间，过期时间到了之后，短链接将失效，无法再访问到原始的长链接。
* 短链接删除：短链接可以删除，删除之后，短链接将失效，无法再访问到原始的长链接。

## 非功能需求

* 高可用：短链接服务需要保证高可用，即使某个节点宕机，也不影响整个服务的正常使用。
* 高性能：短链接服务需要保证高性能，能够支撑每秒十万
* 低延迟：短链接服务需要保证低延迟，用户访问短链接时，能够快速的重定向到原始的长链接。
* 高可扩展性：短链接服务需要保证高可扩展性，能够支持大量的短链接生成和访问。
* 高可靠性：短链接服务需要保证高可靠性，能够保证短链接的生成和访问的正确性。

## 资源预算

* 假设我们的系统每天有 100M 用户在线，即 1亿日活；
* 平均每个用户每天写 0.1 个帖子，为每个帖子生成一个对应的短链接，即每天总共生成 1kw 个短链接：
  * 平均每个短链接-长链接映射关系占用 500Bytes 空间，即每天总共需要 5GB 的存储空间；
  * 每日 10,000,000 次写入操作，1kw/86400s ≈ 116，即平均每秒需要处理 116qps 的写入操作；
  * 假设峰值写入量约平均写入量的 10 倍，即 1160qps，为便于估算，可理解写峰值 qps 为 1k；
* 平均每个用户每天访问 10 个帖子，即每天总共访问 10亿 次短链接，读写比例为 100:1：
  * 平均每秒需要处理 11600 的读取操作，即约为 10k/s；
  * 假设峰值读取量约平均读取量的 10 倍，即约为 100k/s；
* 缓存资源预算：
  * 由于短链服务具有明显的热点数据特征，因此需要使用缓存来提高访问性能，我们假设 10% 的数据贡献了 99% 的访问量
  * 每日 10亿次的访问量中，我们假设 99% 的访问量是固定在 10% 的热点数据上，即需要缓存 1亿条数据，缓存服务器需要 50GB 内存空间；
  * 再进一步，我们利用本地缓存来缓存最热的 1% 的数据，即需要缓存 1M 条数据，每台 Server 节点需要 500MB 内存空间用于本地缓存；
  * 缓存命中率为 99%，即每日 10亿次的访问量中，有 1% 的访问量需要访问数据库，即 10M 次/日，即每秒需要处理 116 次数据库读取操作，即约为 100qps；

综上所述：
  * 数据库存储空间：每日消耗 5GB 磁盘，三年时间需要约 5TB；
  * 数据库读写请求数：平均每秒 116qps 的写入操作与读取操作，峰值 1k/qps 的写入操作与读取操作；
  * 缓存服务器内存空间：总共需要 50GB 内存空间，缓存约 1亿条数据；
  * 本地缓存存储空间：每台 Server 节点需要 500MB 内存空间用于本地缓存，缓存约 1M 条数据；

## 更多设计细节

* [短链接服务系统设计](docs/system-design.md)
* [Base58 编码算法](docs/base58-design.md)
* [分布式 ID 生成器](docs/tddl-design.md)
* [限流器设计](docs/rate-limiter-design.md)
* [API 性能测试](docs/api-benchmark.md)
* [数据库表结构](docs/ddl)