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