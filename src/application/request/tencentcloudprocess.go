package request

import (
	"AutoCert/src/utils"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

// ProcessTencentCloudCertificates handles the SSL certificate application process for TencentCloud domains
func ProcessTencentCloudCertificates(config utils.Config) {
	log.Println("[INFO] Starting TencentCloud certificate processing")

	// Get domains that use TencentCloud for certificate requests
	domains := getDomainsByRequestPlatform(config, "tencentcloud")
	log.Printf("[INFO] Retrieved %d domains for TencentCloud processing", len(domains))

	// Check SSL certificates for all domains
	expiringDomains, expiredDomains, errorDomains := utils.CheckSSLCertificates(config)
	log.Printf("[INFO] Certificate check results: %d expiring, %d expired, %d with errors", len(expiringDomains), len(expiredDomains), len(errorDomains))

	// Combine domains that need certificate renewal
	domainsToRenew := append(expiringDomains, expiredDomains...)
	domainsToRenew = append(domainsToRenew, errorDomains...)

	// Filter domains to only include those that use TencentCloud
	var tencentCloudDomainsToRenew []string
	for _, domain := range domainsToRenew {
		if contains(domains, domain) {
			tencentCloudDomainsToRenew = append(tencentCloudDomainsToRenew, domain)
		}
	}

	log.Printf("[INFO] Applying certificates for %d TencentCloud domains", len(tencentCloudDomainsToRenew))

	var wg sync.WaitGroup

	// Apply for new certificates
	for _, domain := range tencentCloudDomainsToRenew {
		log.Printf("[INFO] Applying for certificate for domain: %s", domain)
		certificateId, err := applyTencentCloudSSLCertificate(domain, config)
		if err != nil {
			log.Printf("[ERROR] Failed to apply for certificate for domain %s: %v", domain, err)
			continue
		}
		log.Printf("[INFO] Certificate application submitted for domain %s, CertificateId: %s", domain, certificateId)

		// Start a goroutine to monitor the certificate status
		wg.Add(1)
		go func(cid string) {
			defer wg.Done()
			monitorCertificateStatus(config, cid)
		}(certificateId)
	}

	// Wait for all goroutines to complete
	log.Println("[INFO] Waiting for all certificate status checks to complete")
	wg.Wait()

	log.Println("[INFO] Completed TencentCloud certificate processing")
}

// monitorCertificateStatus continuously checks the status of a certificate application
func monitorCertificateStatus(config utils.Config, certificateId string) {
	for {
		status, err := DescribeCertificate(config, certificateId)
		if err != nil {
			log.Printf("[ERROR] Failed to describe certificate %s: %v", certificateId, err)
			return
		}

		switch status {
		case 1:
			log.Printf("[INFO] Certificate %s has been approved", certificateId)

			// 调用 getTencentCloudCert 函数
			certFiles, err := getTencentCloudCert(config.TencentCloud.AccessKey, config.TencentCloud.SecretKey, certificateId)
			if err != nil {
				log.Printf("[ERROR] Failed to get certificate files for %s: %v", certificateId, err)
			} else {
				log.Printf("[INFO] Successfully retrieved certificate files for %s", certificateId)
				log.Println("[INFO] Certificate files:")
				for _, file := range certFiles {
					log.Printf("- %s", file)
				}
			}
			return
		case 0, 4:
			log.Printf("[INFO] Certificate %s is still under review", certificateId)
			time.Sleep(30 * time.Second) // Wait for 30 seconds before checking again
		default:
			log.Printf("[ERROR] Unexpected status %d for certificate %s", status, certificateId)
			return
		}
	}
}

// DescribeCertificate checks the status of a certificate application
func DescribeCertificate(config utils.Config, certificateId string) (int64, error) {
	credential := common.NewCredential(
		config.TencentCloud.AccessKey,
		config.TencentCloud.SecretKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"

	client, _ := ssl.NewClient(credential, "", cpf)

	request := ssl.NewDescribeCertificateRequest()
	request.CertificateId = common.StringPtr(certificateId)

	response, err := client.DescribeCertificate(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return 0, fmt.Errorf("API error: %s", sdkError)
		}
		return 0, err
	}

	if response.Response.Status == nil {
		return 0, fmt.Errorf("status is nil")
	}

	return int64(*response.Response.Status), nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
