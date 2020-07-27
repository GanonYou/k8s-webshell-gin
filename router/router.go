package router

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	k8s_ws "k8s-webshell/k8s-ws"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"net/http"
)

var Router *gin.Engine

// ssh流式处理器
type streamHandler struct {
	wsConn      *k8s_ws.WsConnection
	resizeEvent chan remotecommand.TerminalSize
}

// web终端发来的包
type xtermMessage struct {
	MsgType string `json:"type"`  // 类型:resize客户端调整终端, input客户端输入
	Input   string `json:"input"` // msgtype=input情况下使用
	Rows    uint16 `json:"rows"`  // msgtype=resize情况下使用
	Cols    uint16 `json:"cols"`  // msgtype=resize情况下使用
}

// executor回调获取web是否resize
func (handler *streamHandler) Next() (size *remotecommand.TerminalSize) {
	ret := <-handler.resizeEvent
	size = &ret
	return
}

// executor回调读取web端的输入
func (handler *streamHandler) Read(p []byte) (size int, err error) {
	var (
		msg      *k8s_ws.WsMessage
		xtermMsg xtermMessage
	)

	// 读web发来的输入
	if msg, err = handler.wsConn.WsRead(); err != nil {
		return
	}

	// 解析客户端请求
	if err = json.Unmarshal(msg.Data, &xtermMsg); err != nil {
		return
	}

	//web ssh调整了终端大小
	if xtermMsg.MsgType == "resize" {
		// 放到channel里，等remotecommand executor调用我们的Next取走
		handler.resizeEvent <- remotecommand.TerminalSize{Width: xtermMsg.Cols, Height: xtermMsg.Rows}
	} else if xtermMsg.MsgType == "input" { // web ssh终端输入了字符
		// copy到p数组中
		size = len(xtermMsg.Input)
		copy(p, xtermMsg.Input)
	}
	return
}

// executor回调向web端输出
func (handler *streamHandler) Write(p []byte) (size int, err error) {
	var (
		copyData []byte
	)

	// 产生副本
	copyData = make([]byte, len(p))
	copy(copyData, p)
	size = len(p)
	err = handler.wsConn.WsWrite(websocket.TextMessage, copyData)
	return
}


func ProcessK8sWsHtml(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request,"resource/k8s-ws.html")
}

func ProcessXtermCss(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "resource/xterm.css")
}

func ProcessXtermJs(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "resource/xterm.js")
}

func ProcessFitJs(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "resource/fit.js")
}

// func wsHandler(resp http.ResponseWriter,req *http.Request) {
func wsHandler(c *gin.Context) {
	var (
		wsConn    *k8s_ws.WsConnection
		restConf  *rest.Config
		sshReq    *rest.Request
		pod       string
		namespace string
		container string
		executor  remotecommand.Executor
		handler   *streamHandler
		err       error
	)

	// 得到websocket长连接
	if wsConn, err = k8s_ws.InitWebsocket(c.Writer, c.Request); err != nil {
		return
	}

	namespace = c.Query("namespace")
	pod = c.Query("pod")
	container = c.Query("container")

	// 获取k8s rest client配置
	if restConf, err = k8s_ws.GetRestConf(); err != nil {
		goto END
	}

	sshReq = k8s_ws.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   []string{"bash"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	// 创建到容器的连接
	if executor, err = remotecommand.NewSPDYExecutor(restConf, "POST", sshReq.URL()); err != nil {
		goto END
	}

	// 配置与容器之间的数据流处理回调
	handler = &streamHandler{wsConn: wsConn, resizeEvent: make(chan remotecommand.TerminalSize)}
	if err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             handler,
		Stdout:            handler,
		Stderr:            handler,
		TerminalSizeQueue: handler,
		Tty:               true,
	}); err != nil {
		goto END
	}
	return

END:
	fmt.Println(err)
	wsConn.WsClose()
}

// 预留容器SSH鉴权
func CanUserToContainer(c *gin.Context) {
	if c.Query("clusterName") == "qwe" {
		c.Abort()
		c.JSON(http.StatusUnauthorized,"this user cannot ssh this container!")
	}
	c.Next()
}

func ParseK8sWebshellRouter(k8sWsGr *gin.RouterGroup) {
	k8sWsGr.GET("/terminal", ProcessK8sWsHtml)
	k8sWsGr.GET("/xterm.css", ProcessXtermCss)
	k8sWsGr.GET("/xterm.js", ProcessXtermJs)
	k8sWsGr.GET("/fit.js", ProcessFitJs)
	k8sWsGr.GET("/connect",wsHandler)
}

func CreateRouter() {
	Router = gin.Default()
	k8sWsGr := Router.Group("/k8s-ws",CanUserToContainer)
	{
		ParseK8sWebshellRouter(k8sWsGr)
	}
}
