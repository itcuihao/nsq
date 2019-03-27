## nsqlookupd

`nsqlookupd` is the daemon that manages topology metadata and serves client requests to
discover the location of topics at runtime.

Read the [docs](http://nsq.io/components/nsqlookupd.html)


---
参考文章：[NSQ 源码分析之 nsqlookupd](https://juejin.im/entry/5850b63d1b69e6006c75f2c2)

nsqlookupd是nsq管理集群拓扑信息以及用于注册发现的nsqd服务

当nsq集群中有多个nsqlookupd服务时，每个nsqd会向所有的nsqlookupd服务上报本地信息。
所以，nsqlookupd具有最终一致性。

```
apps　目录存放了nsqd, nsqlookupd, nsqadmin和一些工具的main函数文件；
internal　目录存放了NSQ内部使用的一些函数，例如三大组件通用函数；
nsqadmin　目录存放了关于nsqadmin源码；
nsqd　目录存放了关于nsqd源码；
nsqlookupd　目录存放了　nsqlookupd的源码；
```

