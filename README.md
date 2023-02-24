# WSIM

基于 Websocket 实现简单易用的轻量 IM 服务。

IM 服务应该处于业务与客户端之间的独立存在，这里面主要维护房间与用户关系，可以做到广播房间，推送个人，批量推送功能。

* 与客户端交互基于 websocket 协议。
* 与业务的交互通过 grpc 协议。

```s
 ┌───────────┐  websocket ┌───────────┐    grpc    ┌───────────┐
 │           ◄────────────┘           ◄────────────┘           │
 │  CLIENT   ┌────────────►   WSIM    ┌────────────► BUSINESS  │
 └───────────┘            └───────────┘            └───────────┘
```

具体使用可参考 example/http

## 参考项目

* <https://github.com/gorilla/websocket>
* <https://github.com/Terry-Mao/goim>
* <https://github.com/AnupKumarPanwar/Golang-realtime-chat-rooms>
* <https://github.com/grpc/grpc-go>
