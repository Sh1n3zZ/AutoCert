package main

import (
	"AutoCert/src/application/request"
	"AutoCert/src/utils"
	"fmt"
)

func main() {
	config := utils.InitializationConfig()
	utils.CheckSSLCertificates(config)
	fmt.Println("开始处理阿里云SSL证书...")

	request.ProcessAliyunSSLCertificates(config)

	fmt.Println("阿里云SSL证书处理完成")
}
