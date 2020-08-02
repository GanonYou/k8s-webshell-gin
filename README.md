# k8s-webshell-gin

Golang实现登入k8s中指定pod内容器的Webshell功能，基于GIN框架、k8s/client-go，预留组内鉴权中间件

## Quick Start

1. 将项目根目录下的YOUR_

## 流程
- web端GET请求，server端响应前端资源
- webshell发起websocket请求，server端升级连接
- K8s/client-go 建立与container的ssh长连接，通过 websocket 连接实现PtyHandle接口的读写方法。
- 使用组内中间件进行鉴权

具体流程如下图所示:


