package utils

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"
)

func CheckSSLCertificates(config Config) ([]string, []string, []string) {
	log.Println("[INFO] Starting SSL certificate check for all domains")
	expiringDomains := []string{}
	expiredDomains := []string{}
	errorDomains := []string{}

	for _, domain := range config.Domains {
		log.Printf("[INFO] Checking certificate for domain: %s", domain.DomainName)
		domainName, expirationDate, err := checkCertificateExpTime(domain.DomainName)
		if err != nil {
			log.Printf("[ERROR] Failed to check certificate for domain %s: %v", domainName, err)
			errorDomains = append(errorDomains, domainName)
			continue
		}

		timeUntilExpiration := time.Until(expirationDate)
		if timeUntilExpiration <= 0 {
			log.Printf("[WARN] Certificate for domain %s has expired", domainName)
			expiredDomains = append(expiredDomains, domainName)
		} else if timeUntilExpiration.Hours() <= 72 {
			log.Printf("[WARN] Certificate for domain %s is expiring soon (within 72 hours)", domainName)
			expiringDomains = append(expiringDomains, domainName)
		} else {
			log.Printf("[INFO] Certificate for domain %s is valid", domainName)
		}
	}

	log.Println("[INFO] Completed SSL certificate check for all domains")
	return expiringDomains, expiredDomains, errorDomains
}

func checkCertificateExpTime(domain string) (string, time.Time, error) {
	log.Printf("[INFO] Checking certificate expiration time for domain: %s", domain)
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", domain+":443", conf)
	if err != nil {
		return domain, time.Time{}, fmt.Errorf("connection failed: %v", err)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return domain, time.Time{}, fmt.Errorf("SSL certificate not configured")
	}

	cert := state.PeerCertificates[0]
	expirationDate := cert.NotAfter

	// Check certificate chain
	if len(state.PeerCertificates) == 1 {
		log.Printf("[WARN] SSL certificate chain for domain %s is incomplete (single certificate only)", domain)
	}

	// Calculate days, hours, minutes and seconds
	days := int(time.Until(expirationDate).Hours()) / 24
	hours := int(time.Until(expirationDate).Hours()) % 24
	minutes := int(time.Until(expirationDate).Minutes()) % 60
	seconds := int(time.Until(expirationDate).Seconds()) % 60

	// Get local timezone
	localTime := expirationDate.Local()

	// Get GMT+8 timezone
	gmt8, _ := time.LoadLocation("Asia/Shanghai")
	gmt8Time := expirationDate.In(gmt8)

	log.Printf("[INFO] Domain: %s", domain)
	log.Printf("[INFO] Certificate expiration time (local timezone): %s", localTime.Format("2006-01-02 15:04:05 MST"))
	log.Printf("[INFO] Certificate expiration time (GMT+8): %s", gmt8Time.Format("2006-01-02 15:04:05 MST"))
	log.Printf("[INFO] Time until expiration: %d days %d hours %d minutes %d seconds", days, hours, minutes, seconds)

	return domain, expirationDate, nil
}
