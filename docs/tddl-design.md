# TDDL 设计文档

## 概述

TDDL (Tiny Distributed Database Layer) 是一个 Go 语言实现的分布式序列号生成器。它使用 GORM 作为数据库访问层，可以与任何 GORM 支持的数据库一起使用。

## 主要组件

### TDDL 接口

TDDL 接口定义了序列号生成器的主要行为。它有两个方法：

- `Next(ctx context.Context) (uint64, error)`：生成并返回下一个序列号。
- `Close()`：关闭序列号生成器，释放所有资源。

### tddlSequence 结构体

tddlSequence 结构体是 TDDL 接口的一个实现。它使用 GORM 连接、步长、序列名和起始序列号作为初始化参数。

tddlSequence 结构体的主要字段包括：

- `clientID`：客户端 ID，用于区分不同的客户端实例。
- `conn`：GORM 数据库连接。
- `rowID`：序列记录的主键 ID。
- `step`：序列号的步长。
- `max`：当前步长内的最大序列号。
- `curr`：当前步长内的当前序列号。
- `stop`：用于停止 worker 的通道。
- `queue`：用于存储生成的序列号的通道。

### worker

worker 是一个在后台运行的 goroutine，负责生成序列号并将它们发送到 `queue` 通道。当 `curr` 达到 `max` 时，worker 会调用 `renew` 方法更新 `curr` 和 `max`。

### renew 方法

renew 方法负责更新 `curr` 和 `max`。它首先从数据库中获取当前的序列记录，然后使用乐观锁更新该记录的序列号。如果更新成功，它会更新 `curr` 和 `max` 的值。

## 测试

TDDL 的测试主要包括单客户端和多客户端的序列号生成测试，以及超时处理测试。这些测试确保 TDDL 能在各种情况下正确地生成序列号。

## 使用

要使用 TDDL，首先需要创建一个新的 tddlSequence 实例，然后可以调用其 `Next` 方法生成序列号。当不再需要 tddlSequence 时，应调用其 `Close` 方法释放资源。