package request

import (
	"AutoCert/src/utils"
	"fmt"
	"log"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

func createAliDNSClient(config utils.Config) (*alidns20150109.Client, error) {
	clientConfig := &openapi.Config{
		AccessKeyId:     tea.String(config.Aliyun.AccessKey),
		AccessKeySecret: tea.String(config.Aliyun.SecretKey),
	}
	clientConfig.Endpoint = tea.String("alidns.cn-hangzhou.aliyuncs.com")
	return alidns20150109.NewClient(clientConfig)
}

func AddDNSRecord(config utils.Config, domain, recordType, recordDomain, recordValue string) error {
	log.Printf("[INFO] Adding DNS record for domain %s: Type=%s, RR=%s, Value=%s\n", domain, recordType, recordDomain, recordValue)

	client, err := createAliDNSClient(config)
	if err != nil {
		return fmt.Errorf("failed to create AliDNS client: %v", err)
	}

	addDomainRecordRequest := &alidns20150109.AddDomainRecordRequest{
		DomainName: tea.String(domain),
		RR:         tea.String(recordDomain),
		Type:       tea.String(recordType),
		Value:      tea.String(recordValue),
	}

	runtime := &util.RuntimeOptions{}
	_, err = client.AddDomainRecordWithOptions(addDomainRecordRequest, runtime)
	if err != nil {
		return fmt.Errorf("failed to add DNS record: %v", err)
	}

	return nil
}
