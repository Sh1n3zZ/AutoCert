package request

import (
	"AutoCert/src/utils"
	"fmt"
	"log"
	"strconv"
	"time"

	"strings"

	cas20200407 "github.com/alibabacloud-go/cas-20200407/v3/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

func createClient(config utils.Config) (*cas20200407.Client, error) {
	clientConfig := &openapi.Config{
		AccessKeyId:     tea.String(config.Aliyun.AccessKey),
		AccessKeySecret: tea.String(config.Aliyun.SecretKey),
	}
	clientConfig.Endpoint = tea.String("cas.aliyuncs.com")
	return cas20200407.NewClient(clientConfig)
}

func ApplyAliyunSSLCertificate(domain string, config utils.Config) (string, error) {
	client, err := createClient(config)
	if err != nil {
		return "", fmt.Errorf("failed to create Aliyun client: %v", err)
	}

	request := &cas20200407.CreateCertificateForPackageRequestRequest{
		ProductCode:  tea.String("digicert-free-1-free"),
		ValidateType: tea.String("DNS"),
		Domain:       tea.String(domain),
	}
	runtime := &util.RuntimeOptions{}

	log.Printf("[INFO] Applying for Aliyun SSL certificate for domain: %s\n", domain)
	response, err := client.CreateCertificateForPackageRequestWithOptions(request, runtime)
	if err != nil {
		return "", fmt.Errorf("failed to apply for Aliyun SSL certificate: %v", err)
	}

	if response.Body == nil || response.Body.OrderId == nil {
		return "", fmt.Errorf("application successful but no order ID returned")
	}

	orderId := strconv.FormatInt(*response.Body.OrderId, 10)
	log.Printf("[INFO] Successfully applied for Aliyun SSL certificate for domain %s, Order ID: %s\n", domain, orderId)
	return orderId, nil
}

func DescribeAliyunCertificateState(orderId string, config utils.Config, baseDomain string) (string, string, string, string, error) {
	client, err := createClient(config)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to create Aliyun client: %v", err)
	}

	orderIdInt, err := strconv.ParseInt(orderId, 10, 64)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to convert order ID: %v", err)
	}

	request := &cas20200407.DescribeCertificateStateRequest{
		OrderId: tea.Int64(orderIdInt),
	}
	runtime := &util.RuntimeOptions{}

	maxRetries := 3
	retryDelay := time.Second * 5

	var response *cas20200407.DescribeCertificateStateResponse
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		log.Printf("[INFO] Checking certificate status for Order ID: %s (Attempt %d/%d)\n", orderId, i+1, maxRetries)
		response, lastErr = client.DescribeCertificateStateWithOptions(request, runtime)
		if lastErr == nil {
			break
		}
		log.Printf("[WARN] Failed to query certificate status (Attempt %d/%d): %v\n", i+1, maxRetries, lastErr)
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	if lastErr != nil {
		return "", "", "", "", fmt.Errorf("failed to query certificate status after %d attempts: %v", maxRetries, lastErr)
	}

	log.Printf("[INFO] Certificate status query result:\n%s\n", tea.Prettify(response))

	if response.Body == nil {
		return "", "", "", "", fmt.Errorf("empty response body")
	}

	recordDomain := tea.StringValue(response.Body.RecordDomain)
	rr := strings.TrimSuffix(recordDomain, "."+baseDomain)

	log.Printf("[INFO] Order ID: %d\n", orderIdInt)
	log.Printf("[INFO] Certificate Status: %s\n", tea.StringValue(response.Body.Type))
	log.Printf("[INFO] Validation Type: %s\n", tea.StringValue(response.Body.ValidateType))
	log.Printf("[INFO] Record Type: %s\n", tea.StringValue(response.Body.RecordType))
	log.Printf("[INFO] Record Domain: %s\n", recordDomain)
	log.Printf("[INFO] Record Value: %s\n", tea.StringValue(response.Body.RecordValue))
	log.Printf("[INFO] RR: %s\n", rr)
	log.Printf("[INFO] Base Domain: %s\n", baseDomain)

	return tea.StringValue(response.Body.Type),
		tea.StringValue(response.Body.RecordType),
		rr,
		tea.StringValue(response.Body.RecordValue),
		nil
}
