# RateLimiter 设计文档

## 概述

`RateLimiter` 是一个接口，它知道如何限制处理某事的速率。它提供了一些方法来决定一个项目应该等待多长时间，停止跟踪一个项目，以及获取一个项目失败的次数。

## 主要组件

### RateLimiter 接口

`RateLimiter` 接口定义了限制处理速率的主要行为。它有三个方法：

- `When(item T) time.Duration`：获取一个执行对象应该等待多长时间。
- `Forget(item T)`：表示一个执行对象完成了重试。无论是因为失败还是成功，我们都会停止跟踪它。
- `Retries(item T) int`：返回执行对象失败的次数。

这个接口是泛型的，可以接受任何可比较的类型 `T`。

## 实现

`RateLimiter` 接口有多种实现，包括 `BucketRateLimiter`、`ItemExponentialFailureRateLimiter`、`ItemFastSlowRateLimiter` 和 `MaxOfRateLimiter`。这些实现提供了不同的限制策略，包括令牌桶限制、指数退避限制、快慢速率限制和最大速率限制。

每种实现都有自己的特性和使用场景。例如，`BucketRateLimiter` 使用标准的令牌桶进行限制，`ItemExponentialFailureRateLimiter` 使用基于指数退避的限制策略，`ItemFastSlowRateLimiter` 在一定次数的尝试后从快速重试切换到慢速重试，`MaxOfRateLimiter` 则从多个限制器中选择最严格的限制。

## 使用

要使用 `RateLimiter`，首先需要创建一个新的限制器实例，然后可以调用其 `When` 方法获取项目应该等待的时间。当项目完成重试时，应调用其 `Forget` 方法停止跟踪。可以通过 `Retries` 方法获取项目失败的次数。