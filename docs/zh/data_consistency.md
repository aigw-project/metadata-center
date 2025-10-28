# 数据一致性

## 概述

元数据中心实现了垃圾回收（GC）机制来维护数据一致性并防止内存泄漏。该机制自动清理过期的元数据条目，确保负载统计的准确性。

## GC流程

### 1. 周期性GC执行
```go
// GC默认每60秒运行一次
gcInterval = 60 * time.Second
```

### 2. 请求过期
- **默认过期时间**：660秒（11分钟）
- **过期检查**：比较请求创建时间与当前时间
- **清理过程**：移除过期请求并更新统计信息

### 3. 逐步清理

#### 步骤1：清理过期请求
```go
ls.Requests.Range(func(key, value any) bool {
    req := value.(*InferenceRequest)
    if req.CreateTime.Add(requestExpireDuration).Before(now) {
        ls.Requests.Delete(key)
        ls.decEngineStats(req)  // 更新引擎统计信息
    }
    return true
})
```

#### 步骤2：清理过期模型
```go
ls.RunningModelStats.Range(func(key, value any) bool {
    modelStats := value.(*ModelStats)
    if nowStamps >= modelStats.UpdateTime+expire {
        ls.RunningModelStats.Delete(key)
        modelStats.MetricClean()
    }
    return true
})
```

#### 步骤3：清理过期引擎
```go
modelStats.Engines.Range(func(k, v any) bool {
    engineStats := v.(*EngineStats)
    if nowStamps >= engineStats.UpdatedTime+expire {
        modelStats.Delete(k.(string))
        engineStats.MetricClean(key.(string))
    }
    return true
})
```

## 并发处理

### 1. 线程安全操作
- 使用 `sync.Map` 进行并发访问
- 统计信息更新的原子操作

### 2. 延迟重试机制
```go
// 处理删除发生在添加之前的竞态条件
time.AfterFunc(time.Second, func() {
    if !ls.tryDeleteRequestStats(requestID) {
        logger.Warnf("延迟后仍未找到请求ID")
        return
    }
})
```

## 配置参数
### 参数说明

| 参数 | 环境变量 | 描述 | 默认值 |
|------|----------|------|--------|
| `gcInterval` | `METADATA_CENTER_LOAD_GC_INTERVAL` | GC执行间隔 | 60秒 |
| `requestExpireDuration` | `METADATA_CENTER_LOAD_REQ_EXPIRE` | 请求过期时间 | 660秒 |

### 配置调优
- 根据负载和内存约束调整 `gcInterval`
- 根据请求生命周期设置 `requestExpireDuration`

## 数据一致性保证

### 1. 最终一致性
- GC确保陈旧数据最终被移除
- 统计信息在过期窗口内保持准确

### 2. 内存管理
- 防止无限制的内存增长
- 自动清理孤立的条目
- 通过周期性GC实现高效内存使用