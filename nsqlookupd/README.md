## nsqlookupd

`nsqlookupd` is the daemon that manages topology metadata and serves client requests to
discover the location of topics at runtime.

Read the [docs](http://nsq.io/components/nsqlookupd.html)


---

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

```
NSQ提供了三大组件以及一些工具，三大组件为:

nqsd NSQ主要组件，用于存储消息以及分发消息；
nsqlookupd 用于管理nsqd集群拓扑，提供查询nsqd主机地址的服务以及服务最终一致性；
nsqadmin 用于管理以及查看集群中的topic,channel,node等等；

简化一下：
nsqd：队列数据存储
nsqlookup：管理nsqd节点，服务发现
nsqadmin：nsq的可视化
```

```
NSQ的拓扑结构和文件系统的拓扑结构类似，有一个中心节点来管理集群节点；我们从图中可以看出:

nsqlookupd服务同时开启tcp和http两个监听服务，nsqd会作为客户端，连上nsqlookupd的tcp服务，并上报自己的topic和channel信息，以及通过心跳机制判断nsqd状态；还有个http服务提供给nsqadmin获取集群信息；
nsqadmin只开启http服务，其实就是一个web服务，提供给客户端查询集群信息；
nsqd也会同时开启tcp和http服务，两个服务都可以提供给生产者和消费者，http服务还提供给nsqadmin获取该nsqd本地topic和channel信息；
```

**参考文章：**

[NSQ 源码分析之 nsqlookupd](http://luodw.cc/2016/12/13/nsqlookupd/)
