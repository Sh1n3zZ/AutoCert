package request

import (
	"AutoCert/src/utils"
	"log"
	"sync"
	"time"
)

func ProcessAliyunSSLCertificates(config utils.Config) {
	// Get all domains that need to apply for certificates using Aliyun
	aliyunDomains := GetDomainsByRequestPlatform(config, "aliyun")
	log.Println("[INFO] Retrieved domains for Aliyun certificate application:", aliyunDomains)

	// Check SSL certificate status
	expiringDomains, expiredDomains, errorDomains := utils.CheckSSLCertificates(config)
	log.Println("[INFO] Checked SSL certificates status")

	// Merge domains that need to be processed
	domainsToProcess := append(expiringDomains, expiredDomains...)
	domainsToProcess = append(domainsToProcess, errorDomains...)
	log.Println("[INFO] Domains to process:", domainsToProcess)

	// Filter out domains that need to apply using Aliyun
	var aliyunDomainsToProcess []string
	for _, domain := range domainsToProcess {
		for _, aliyunDomain := range aliyunDomains {
			if domain == aliyunDomain {
				aliyunDomainsToProcess = append(aliyunDomainsToProcess, domain)
				break
			}
		}
	}
	log.Println("[INFO] Aliyun domains to process:", aliyunDomainsToProcess)

	// Apply for certificates for each domain and check status
	for _, domainConfig := range config.Domains {
		if domainConfig.RequestPlatform != "aliyun" {
			continue
		}

		domain := domainConfig.DomainName
		baseDomain := domainConfig.BaseDomain

		orderId, err := ApplyAliyunSSLCertificate(domain, config)
		if err != nil {
			log.Printf("[ERROR] Failed to apply certificate for domain %s: %v\n", domain, err)
			continue
		}

		log.Printf("[INFO] Successfully applied certificate for domain %s, Order ID: %s\n", domain, orderId)

		// Wait for a while before checking the certificate status
		time.Sleep(30 * time.Second)

		status, recordType, rr, recordValue, err := DescribeAliyunCertificateState(orderId, config, baseDomain)
		if err != nil {
			log.Printf("[ERROR] Failed to check certificate status for domain %s: %v\n", domain, err)
			continue
		}

		if status == "domain_verify" {
			// Add DNS record
			err = AddDNSRecord(config, baseDomain, recordType, rr, recordValue)
			if err != nil {
				log.Printf("[ERROR] Failed to add DNS record for domain %s. Manual operation required:\n", domain)
				log.Printf("Domain: %s\nRecord Type: %s\nRR: %s\nRecord Value: %s\n", baseDomain, recordType, rr, recordValue)
				continue
			}
			log.Printf("[INFO] Successfully added DNS record for domain %s\n", baseDomain)

			// Start a goroutine to continuously check the certificate status
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					status, _, _, _, err := DescribeAliyunCertificateState(orderId, config, baseDomain)
					if err != nil {
						log.Printf("[ERROR] Failed to check certificate status for domain %s: %v\n", domain, err)
						return
					}

					if status != "domain_verify" && status != "process" {
						log.Printf("[INFO] Certificate status for domain %s: %s\n", domain, status)
						if status == "certificate" {
							log.Printf("[INFO] Certificate issued for domain %s\n", domain)
							// Here you might want to retrieve and log the actual certificate and private key
						}
						return
					}

					time.Sleep(60 * time.Second) // Wait for 60 seconds before checking again
				}
			}()

			wg.Wait() // Wait for the goroutine to finish
		} else {
			log.Printf("[INFO] Certificate status for domain %s is %s, no DNS verification needed\n", domain, status)
		}
	}
}
