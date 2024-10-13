package request

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"archive/zip"
	"bytes"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

func getTencentCloudCert(secretId, secretKey, certificateId string) ([]string, error) {
	// Initialize authentication object
	credential := common.NewCredential(secretId, secretKey)
	log.Println("[INFO] Initialized authentication object")

	// Initialize client options
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
	log.Println("[INFO] Initialized client options")

	// Initialize client object
	client, _ := ssl.NewClient(credential, "", cpf)
	log.Println("[INFO] Initialized client object")

	// Initialize request object
	request := ssl.NewDescribeDownloadCertificateUrlRequest()
	request.CertificateId = common.StringPtr(certificateId)
	request.ServiceType = common.StringPtr("nginx")
	log.Println("[INFO] Initialized request object")

	// Send request
	response, err := client.DescribeDownloadCertificateUrl(request)
	if err != nil {
		log.Printf("[ERROR] Failed to get download URL: %v", err)
		return nil, fmt.Errorf("failed to get download URL: %v", err)
	}
	log.Println("[INFO] Successfully got download URL")

	// Download certificate
	resp, err := http.Get(*response.Response.DownloadCertificateUrl)
	if err != nil {
		log.Printf("[ERROR] Failed to download certificate: %v", err)
		return nil, fmt.Errorf("failed to download certificate: %v", err)
	}
	defer resp.Body.Close()

	zipContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	log.Println("[INFO] Successfully downloaded certificate")

	// Create target directory in the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("[ERROR] Failed to get current working directory: %v", err)
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}
	targetDir := filepath.Join(currentDir, "gitignore", "tencentcloudssl")
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		log.Printf("[ERROR] Failed to create target directory: %v", err)
		return nil, fmt.Errorf("failed to create target directory: %v", err)
	}
	log.Printf("[INFO] Created target directory: %s", targetDir)

	// Save the zip file with the name from SDK response
	zipFileName := *response.Response.DownloadFilename
	zipFilePath := filepath.Join(targetDir, zipFileName)
	err = ioutil.WriteFile(zipFilePath, zipContent, 0644)
	if err != nil {
		log.Printf("[ERROR] Failed to save zip file: %v", err)
		return nil, fmt.Errorf("failed to save zip file: %v", err)
	}
	log.Printf("[INFO] Saved zip file to: %s", zipFilePath)

	// Unzip file
	zipReader, err := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
	if err != nil {
		log.Printf("[ERROR] Failed to read zip content: %v", err)
		return nil, fmt.Errorf("failed to read zip content: %v", err)
	}
	log.Println("[INFO] Successfully read zip content")

	var extractedFiles []string

	for _, file := range zipReader.File {
		filePath := filepath.Join(targetDir, file.Name)
		extractedFiles = append(extractedFiles, filePath)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, 0755)
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			log.Printf("[ERROR] Failed to open zip file: %v", err)
			return nil, fmt.Errorf("failed to open zip file: %v", err)
		}

		targetFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			log.Printf("[ERROR] Failed to create target file: %v", err)
			fileReader.Close()
			return nil, fmt.Errorf("failed to create target file: %v", err)
		}

		_, err = io.Copy(targetFile, fileReader)
		if err != nil {
			log.Printf("[ERROR] Failed to write file content: %v", err)
			fileReader.Close()
			targetFile.Close()
			return nil, fmt.Errorf("failed to write file content: %v", err)
		}

		fileReader.Close()
		targetFile.Close()

		log.Printf("[INFO] Extracted file: %s", filePath)
	}

	log.Printf("[INFO] Successfully extracted %d files", len(extractedFiles))
	return extractedFiles, nil
}
