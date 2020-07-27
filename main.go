package main

import (
	"fmt"
	k8s_ws "k8s-webshell/k8s-ws"
	"k8s-webshell/router"
)

func main() {

	var err error

	// 创建k8s客户端
	if k8s_ws.ClientSet, err = k8s_ws.InitClient(); err != nil {
		fmt.Println(err)
		return
	}

	router.CreateRouter()
	router.Router.Run(":8888")
}
