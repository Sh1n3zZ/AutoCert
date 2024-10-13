package main

import (
	"AutoCert/src/application/request"
	"AutoCert/src/utils"
	"log"
)

func main() {
	config := utils.InitializationConfig()
	utils.CheckSSLCertificates(config)

	// fmt.Println("开始处理阿里云SSL证书...")
	// request.ProcessAliyunSSLCertificates(config)
	// fmt.Println("阿里云SSL证书处理完成")

	log.Println("[INFO] Starting TencentCloud SSL certificate processing")
	request.ProcessTencentCloudCertificates(config)
	log.Println("[INFO] TencentCloud SSL certificate processing completed")
}
