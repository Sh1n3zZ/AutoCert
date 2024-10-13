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

func applyTencentCloudSSLCertificate(domain string, config utils.Config) (string, error) {
	log.Printf("[INFO] Starting SSL certificate application for domain: %s", domain)

	// Instantiate the authentication object using SecretId and SecretKey read from config.toml
	credential := common.NewCredential(
		config.TencentCloud.AccessKey,
		config.TencentCloud.SecretKey,
	)
	log.Println("[DEBUG] Credential object created")

	// Instantiate client options
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
	log.Println("[DEBUG] Client profile created")

	// Instantiate the client object for the product to be requested, clientProfile is optional
	client, err := ssl.NewClient(credential, "", cpf)
	if err != nil {
		log.Printf("[ERROR] Failed to create SSL client: %v", err)
		return "", fmt.Errorf("failed to create SSL client: %v", err)
	}
	log.Println("[DEBUG] SSL client created successfully")

	// Instantiate a request object, each interface will correspond to a request object
	request := ssl.NewApplyCertificateRequest()
	request.DvAuthMethod = common.StringPtr("DNS_AUTO")
	request.DomainName = common.StringPtr(domain)
	log.Println("[DEBUG] Certificate request object created")

	// Send certificate application request
	log.Println("[INFO] Sending certificate application request")
	response, err := client.ApplyCertificate(request)
	if err != nil {
		if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
			log.Printf("[ERROR] An API error has returned: %s", sdkErr)
			return "", fmt.Errorf("API error: %s", sdkErr)
		}
		log.Printf("[ERROR] Unexpected error occurred: %v", err)
		return "", fmt.Errorf("unexpected error: %v", err)
	}

	// Output the response as a JSON string
	log.Println("[INFO] Certificate application successful")
	log.Printf("[DEBUG] Response: %s", response.ToJsonString())

	if response.Response.CertificateId == nil {
		return "", fmt.Errorf("CertificateId is nil in the response")
	}

	return *response.Response.CertificateId, nil
}
