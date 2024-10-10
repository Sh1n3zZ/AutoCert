package request

import (
	"AutoCert/src/utils"
	"fmt"
	"log"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

// GetTencentCloudDomains 获取所有申请平台为 tencentcloud 的域名
func GetTencentCloudDomains(config utils.Config) []string {
	var domains []string
	for _, domain := range config.Domains {
		if domain.RequestPlatform == "tencentcloud" {
			domains = append(domains, domain.DomainName)
		}
	}
	return domains
}

func ApplyTencentCloudSSLCertificate(domain string, config utils.Config) {
	// 实例化认证对象，使用从 config.toml 读取的 SecretId 和 SecretKey
	credential := common.NewCredential(
		config.TencentCloud.AccessKey,
		config.TencentCloud.SecretKey,
	)
	// 实例化client选项
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"

	// 实例化要请求产品的client对象,clientProfile是可选的
	client, err := ssl.NewClient(credential, "", cpf)
	if err != nil {
		log.Fatalf("Failed to create SSL client: %v", err)
	}

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := ssl.NewApplyCertificateRequest()
	request.DvAuthMethod = common.StringPtr("DNS_AUTO")
	request.DomainName = common.StringPtr(domain)

	// 发送申请证书请求
	response, err := client.ApplyCertificate(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return
	}
	if err != nil {
		panic(err)
	}

	// 输出json格式的字符串回包
	fmt.Printf("%s\n", response.ToJsonString())
}
