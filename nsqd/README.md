## nsqd

`nsqd` is the daemon that receives, queues, and delivers messages to clients.

Read the [docs](http://nsq.io/components/nsqd.html)


**消息发送者PUB流程**

```
生产者PUB topic消息；
topic.PutMessage 到 topic.memoryMsgChan；
每topic消息有后台协程topic.messagePump进行Topic.memoryMsgChan监听；
topic.messagePump 遍历每个channel调用channel.PutMessage发送消息到channel.memoryMsgChan ;
```

**发送者总结**

```
1. nsqd每个连接上来的客户端有一个协程IOLoop循环读取消息，还有一个协程负责订阅channel的消息管道；

2. 每一个topic有一个后台协程进行消息的发送和其他处理;

3. 客户端发送消息时是在IOLoop里面把消息放入topic.memoryMsgChan的，之后就由这个topic的messagePump来处理消息；

4. 每个channel没有单独的协程负责后面的处理，channel发送消息其实就只需要把消息放到channel.memoryMsgChan就行；
```

**消息订阅者SUB流程**

```
1.消费者使用TCP协议，发送SUB topic channel 命令订阅到某个channel上，记录其client.Channel = channel，通知c.SubEventChan；
2.消费者启动后台协程protocolV2.messagePump订阅c.SubEventChan 并得知channel有订阅消息；
3.开始订阅管道subChannel.memoryMsgChan/backend， 每个客户端都可以订阅到channel的内存或者磁盘队列里面;
4.待生产者调用第四步后，其中一个client会得到消息：msg := <-memoryMsgChan;
5.客户端记录StartInFlightTimeout的发送中消息队列，进行超时处理；
6.SendMessage 将消息+msgid发送给消费者；
7.消费者收到msgid后，发送FIN+msgid通知服务器成功投递消息，可以清空消息了；
```
**总结订阅者**
```
1. nsqd每个连接上来的客户端会创建一个protocolV2.messagePump协程负责订阅消息，做超时处理；

2. 客户端发送SUB命令后，SUB()函数会通知客户端的messagePump携程去订阅这个channel的消息；

3. messagePump协程收到订阅更新的管道消息后，会等待在Channel.memoryMsgChan和Channel.backend.ReadChan()上；

4. 只要有生产者发送消息后，channel.memoryMsgChan便会有新的消息到来，其中一个客户端就能获得管道的消息；

5. 拿到消息后调用StartInFlightTimeout将消息放到队列，用来做超时重传，然后调用SendMessage发送给客户端 ；

消费者流程基本就是这样了，稍微要注意的就是nsq的消息有defer和超时重传的功能，两个功能分别由StartDeferredTimeout 和StartInFlightTimeout 两个优先级队列来维护，后面对接的其实是Main() 函数创建的queueScanLoop队列scan扫描协程；
```