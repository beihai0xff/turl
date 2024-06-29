
# 系统设计

## 业务流程

### 短链接生成

1. 用户输入长链接，点击生成短链接；
2. 服务端接收到请求，从发号器获取唯一的标识 ID，并将 UID 转换为 8位 Base58 编码；
3. 服务端将短链接与长链接的映射关系存储到数据库中；

### 短链接访问

1. 用户访问短链接；
2. 服务端接收到请求，从短链接中解析出 UID；
3. 服务端根据 UID 依次从本地缓存、远程缓存、数据库中获取长链接；
4. 服务端返回 302 临时重定向，将用户重定向到原始的长链接；

## 技术选型

### 分布式 ID

短链接发号器需要保证生成的短链接是唯一的，可以使用分布式 ID 生成器，如 Twitter 的 Snowflake 算法，或者使用数据库自增 ID 生成器等等，下面将描述不同方案的优缺点：

**数据库自增 ID**

将数据库自增 ID 作为短链接 UID，然后将 UID 转换为 Base58 编码，即可生成短链接。
* 优点：简单易用，生成的 ID 是唯一的；
* 缺点：依赖于数据库，数据库的性能将成为瓶颈，不适合高并发场景，如果采用数据库集群，需要避免 ID 重复；并且数据库自增 ID 是有序的，可能会暴露业务规模；

**Redis 自增序列**

将 Redis 的自增序列作为短链接 UID，然后将 UID 转换为 Base58 编码，即可生成短链接。
* 优点：高性能，生成的 ID 是唯一的；
* 缺点：Redis 是基于内存的，如果 Redis 宕机，可能会导致 ID 重复，需要保证 Redis 的高可用；ID 自增是有序的，可能会暴露业务规模；

**UUID**

使用 UUID 作为短链接 UID。
* 优点：生成的 ID 是唯一的，不依赖于数据库；
* 缺点：UUID 是 128 位的，转换为 Base58 编码后，短链接长度过长，不适合短链接服务；且 UUID 是无序的，插入数据时可能会导致数据库的性能问题；

**Snowflake 算法**
* 优点：高性能，高可用，生成的 ID 是唯一的；
* 缺点：需要依赖于时钟，时钟回拨会导致 ID 重复，需要保证时钟的稳定性；

**TDDL 序列**

在数据库中创建一个序列表，用于存储序列的当前值，然后通过数据库的原子操作来获取下一个序列区间，然后在内存中递增序列值，当序列值用尽时，再次获取下一个序列区间。
* 优点：生成的 ID 是唯一的，避免了数据库自增 ID 的性能瓶颈；
* 缺点：具有一定的维护成本。

综上所述，我们可以选择 TDDL 序列算法作为短链接发号器，保证生成的短链接是唯一的，同时 TDDL 序列算法也便于根据 short ID 进行分库分表，适合高并发场景。

### 数据存储

短链接服务需要存储短链接与长链接的映射关系，可以选择关系型数据库、NoSQL 数据库等，turl 首要支持 MySQL 等关系型数据库，未来考虑支持 MongoDB 等 NoSQL 数据库。

MySQL 数据库表设计如下：

```sql
create table turl.tiny_urls
(
    id         bigint unsigned auto_increment
        primary key,
    created_at datetime(3)  null,
    updated_at datetime(3)  null,
    deleted_at datetime(3)  null,
    long_url   varchar(500) not null,
    short      bigint       not null
);

create index idx_tiny_urls_deleted_at
    on turl.tiny_urls (deleted_at);

create index idx_tiny_urls_short
    on turl.tiny_urls (short);


```



