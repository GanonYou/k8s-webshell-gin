package k8s_ws

import (
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var ClientSet *kubernetes.Clientset

// 初始化k8s客户端
func InitClient() (*kubernetes.Clientset, error) {

	var restConf *rest.Config
	var clientSet *kubernetes.Clientset
	var err error

	if restConf, err = GetRestConf(); err != nil {
		return nil, err
	}
	// 生成client set配置
	if clientSet, err = kubernetes.NewForConfig(restConf); err != nil {
		return nil, err
	}

	return clientSet, nil
}

// 获取k8s restful client配置
func GetRestConf() (*rest.Config, error) {

	var restConf *rest.Config
	var err error
	var kubeConfig []byte

	// 读kubeConfig文件
	if kubeConfig, err = ioutil.ReadFile("./YOUR_K8S.conf"); err != nil {
		return nil, err
	}
	// 生成rest client配置
	if restConf, err = clientcmd.RESTConfigFromKubeConfig(kubeConfig); err != nil {
		return nil, err
	}

	return restConf, nil
}
