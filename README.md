[![GoDoc](https://godoc.org/github.com/CJey/gbase?status.svg)](https://godoc.org/github.com/CJey/gbase)

## gbase

目标在于实现非常基础的常用库，提供简单易用的API

```go
import "github.com/cjey/gbase"
```

### env

检查当前版本信息

```bash
$ ./env.sh
```

查看changelog

```bash
$ ./changelog.sh
```

### doc

主包会自动集成或者export子包的常用形式

查看完整的文档

```bash
# 主包
go doc -all github.com/cjey/gbase

# Context
go doc -all github.com/cjey/gbase/context
```

### Context

Context在实现了官方的Context接口外，额外追加了环境变量和日志能力，主包在此基础上做了个应用封装

#### 关键特性

**兼容**

实现了官方库的Context接口

**WithX**

一组操作，用于直接简单的执行Context派生

**Fork**

Context支持执行fork操作，将会copy当前的环境，生成新的子级环境，同时其name后会自动的被追加一个数字后缀(location保持)

一般使用场景：一次请求当中，需要开启并发操作，并发的任务当中使用fork的Context来执行操作，这样在日志输出上就可以简单的区分出子任务

如果本请求触发了一个异步任务，则需要谨慎对待，因为派生Context用于异步可能会存在副作用（常规现象，继承的Context很可能会在同步请求结束时被立即Cancel），解决的办法可以是根据情况创建一个新的NamedContext，手工继承源Context的Name和Location(按需)

**At**

此操作将会copy当前的环境，生成新的子级环境，同时会在原有的Location之上合并入一个新的location名称(name保持)，Context支持在新的逻辑过程中自行指明一个位置名称，在日志输出时便会自动带上该位置，并且，此位置名称是会被继承的，这样在输出日志中便可以简单的查看到一定的逻辑调用层次关系

当然，这个是可选的，也不建议在大大小小所有的函数位置处使用此能力，否则跟每条日志都打印一次调用堆栈没了区别，应当根据情况用在合适的位置

**Set/Get**

这是一组快捷操作，等价于调用Context内的Env，用于提供类似环境变量的能力，每当有新的Context派生出来之时，即会同步派生出新的变量空间

Set操作将只针对自己当前的变量空间，不会影响上游空间，而Get操作则会优先访问当前的变量空间，如果没有找到，则会逐级向上游追溯

**Debug/Info/...**

这是一组快捷操作，等价于调用Context内的Logger，用于提供基本的日志操作，内部的logger选择了zap.SugaredLogger，日志格式也默认被重新调整过，如果想定制格式，可以直接自行执行全局替换

如果只是想要在默认风格上调整日志等级/日志文件/日志编码等配置，可以直接使用ReplaceZapLogger完成目标

默认：ReplaceZapLogger("debug", "stderr", "console", false)

info+json: ReplaceZapLogger("info", "stderr", "json", false)

#### NamedContext

Context可以支持自命名，一般用于伴随服务生命周期的任务，或者是定时执行类的任务

对于这样的用法，需要注意的是，Context创建时，其内部的logger已经固定，比如定时执行类的任务，如果在运行时需要动态调整logger，应当在每次执行时重新创建一个同名的Context

```go
var ctx = gbase.NamedContext("ZookeeperMonitor")
ctx.Info("started")
// console encode
// 2020-04-11 21:24:45.361  info    ZookeeperMonitor started
// json encode
// {"L":"info","T":"2020-04-11 21:24:45.361","N":"ZookeeperMonitor","M":"started"}

ctx = ctx.At("GetID")
ctx.Set("id", "123")
ctx.Info("got it!", "id", ctx.GetString("id"))
// console encode
// 2020-04-11 21:24:45.361  info    ZookeeperMonitor got it! {"@": "GetID", "id": "123"}
// json encode
// {"L":"info","T":"2020-04-11 21:24:45.361","N":"ZookeeperMonitor","M":"got it!", "@": "GetID", "id": "123"}
```

#### SessionContext

还可以选择使用自动生成uuid风格名称的Context，适用于服务请求开始时，为本次请求创建一个SessionContext。

生成的Context的name即是一个带有一定规律的uuid（以BootID的前24位为前缀+12位10进制数字递增序列）

你还可以通过替换变量SessionNameGenerator的方式来实现定制session的生成规则

```go
var ctx = gbase.SessionContext()
ctx.Info("started")
// console encode
// 2020-04-11 21:24:45.361  info    9b2119d3-7f37-4033-8c19-000000000001 started
// json encode
// {"L":"info","T":"2020-04-11 21:24:45.361","N":"9b2119d3-7f37-4033-8c19-000000000001","M":"started"}

ctx = ctx.ForkAt("AsyncGetID")
ctx.Info("show default session", "session", GetSesson(ctx))
// console encode
// 2020-04-11 21:24:45.361  info    9b2119d3-7f37-4033-8c19-000000000001.1 show default session {"@": "AsyncGetID", "session": "9b2119d3-7f37-4033-8c19-000000000001.1"}
// json encode
// {"L":"info","T":"2020-04-11 21:24:45.361","N":"9b2119d3-7f37-4033-8c19-000000000001.1","M":"show default session", "@": "AsyncGetID", "session": "9b2119d3-7f37-4033-8c19-000000000001.1"}

SetSession(ctx, "7ec17674-1360-4fb1-9245-bd8d8d5866c4")
ctx.Info("use my own session", "session", GetSession(ctx))
// console encode
// 2020-04-11 21:24:45.361  info    9b2119d3-7f37-4033-8c19-000000000001.1 use my own session {"@": "AsyncGetID", "session": "7ec17674-1360-4fb1-9245-bd8d8d5866c4"}
// json encode
// {"L":"info","T":"2020-04-11 21:24:45.361","N":"9b2119d3-7f37-4033-8c19-000000000001.1","M":"use my own session", "@": "AsyncGetID", "session": "7ec17674-1360-4fb1-9245-bd8d8d5866c4"}
```

#### ToContext

NamedContext和SessionContext都使用官方context的Background作为内部context，如果需要使用自定义的context或者将官方的context执行转换，则可以使用**ToSessionContext**和**ToNamedContext**
