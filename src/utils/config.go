package utils

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Aliyun struct {
		AccessKey string `toml:"access_key"`
		SecretKey string `toml:"secret_key"`
	} `toml:"aliyun"`

	TencentCloud struct {
		AccessKey string `toml:"access_key"`
		SecretKey string `toml:"secret_key"`
	} `toml:"tencentcloud"`

	AkiLight struct {
		AccessKey string `toml:"access_key"`
		SecretKey string `toml:"secret_key"`
		Endpoint  string `toml:"endpoint"`
	} `toml:"akilight"`

	Domains []Domain `toml:"domains"`
}

type Domain struct {
	DomainName      string `toml:"domain_name"`
	BaseDomain      string `toml:"base_domain"`
	RequestPlatform string `toml:"request_platform"`
	DeployPlatform  string `toml:"deploy_platform"`
}

func InitializationConfig() Config {
	var config Config

	// 读取并解析 config.toml 文件
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	// 打印读取到的 AK 和 SK
	fmt.Println("阿里云 Access Key:", config.Aliyun.AccessKey)
	fmt.Println("阿里云 Secret Key:", config.Aliyun.SecretKey)
	fmt.Println("腾讯云 Access Key:", config.TencentCloud.AccessKey)
	fmt.Println("腾讯云 Secret Key:", config.TencentCloud.SecretKey)

	// 打印域名信息
	for _, domain := range config.Domains {
		fmt.Printf("域名: %s, 申请平台: %s, 部署平台: %s\n",
			domain.DomainName, domain.RequestPlatform, domain.DeployPlatform)
	}

	return config
}
