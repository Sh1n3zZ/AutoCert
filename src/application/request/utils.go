package request

import "AutoCert/src/utils"

func GetDomainsByRequestPlatform(config utils.Config, requestPlatform string) []string {
	var domains []string
	for _, domain := range config.Domains {
		if domain.RequestPlatform == requestPlatform {
			domains = append(domains, domain.DomainName)
		}
	}
	return domains
}
