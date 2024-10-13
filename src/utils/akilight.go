package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type AccessTokenResponse struct {
	Code int `json:"code"`
	Data struct {
		Token     string `json:"token"`
		ExpiresAt int64  `json:"expiresAt"`
	} `json:"data"`
	Message string `json:"message"`
}

var (
	accessToken     string
	accessTokenLock sync.RWMutex
	expiresAt       int64
	cachedToken     string
	cachedExpiresAt int64
	tokenMutex      sync.RWMutex
)

func getAccessToken(config *Config) (string, error) {
	accessTokenLock.RLock()
	if accessToken != "" && time.Now().Unix() < expiresAt {
		fmt.Println("[INFO] Using existing valid access token")
		defer accessTokenLock.RUnlock()
		return accessToken, nil
	}
	accessTokenLock.RUnlock()

	accessTokenLock.Lock()
	defer accessTokenLock.Unlock()

	// Double check to prevent token update while waiting for the lock
	if accessToken != "" && time.Now().Unix() < expiresAt {
		fmt.Println("[INFO] Using existing valid access token after lock acquisition")
		return accessToken, nil
	}

	fmt.Println("[INFO] Requesting new access token")
	url := fmt.Sprintf("%s/APIAccessTokenService/getAPIAccessToken", config.AkiLight.Endpoint)
	requestBody, _ := json.Marshal(map[string]string{
		"type":        "admin",
		"accessKeyId": config.AkiLight.AccessKey,
		"accessKey":   config.AkiLight.SecretKey,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("[ERROR] Failed to request AccessToken: %v\n", err)
		return "", fmt.Errorf("[ERROR] Failed to request AccessToken: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[ERROR] Failed to read response: %v\n", err)
		return "", fmt.Errorf("[ERROR] Failed to read response: %v", err)
	}

	var tokenResp AccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		fmt.Printf("[ERROR] Failed to parse response: %v\n", err)
		return "", fmt.Errorf("[ERROR] Failed to parse response: %v", err)
	}

	if tokenResp.Code != 200 {
		fmt.Printf("[ERROR] Failed to get AccessToken: %s\n", tokenResp.Message)
		return "", fmt.Errorf("[ERROR] Failed to get AccessToken: %s", tokenResp.Message)
	}

	accessToken = tokenResp.Data.Token
	expiresAt = tokenResp.Data.ExpiresAt
	fmt.Println("[INFO] Successfully obtained new access token")

	return accessToken, nil
}

// 使用以下方式以调用已缓存的 token 信息
//
// token, err := getCachedToken(config)
// if err != nil {
// 	return err
// }

func GetCachedToken(config *Config) (string, error) {
	tokenMutex.RLock()
	if cachedToken != "" && time.Now().Unix() < cachedExpiresAt {
		fmt.Println("[INFO] Using cached token")
		defer tokenMutex.RUnlock()
		return cachedToken, nil
	}
	tokenMutex.RUnlock()

	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	// Check again to prevent token update while waiting for the lock
	if cachedToken != "" && time.Now().Unix() < cachedExpiresAt {
		fmt.Println("[INFO] Using cached token after lock acquisition")
		return cachedToken, nil
	}

	fmt.Println("[INFO] Cached token expired or not available, requesting new token")
	// Call GetAccessToken to get a new token
	newToken, err := getAccessToken(config)
	if err != nil {
		fmt.Printf("[ERROR] Failed to get new access token: %v\n", err)
		return "", err
	}

	cachedToken = newToken
	cachedExpiresAt = expiresAt // Use the expiration time set in GetAccessToken
	fmt.Println("[INFO] Successfully updated cached token")

	return cachedToken, nil
}
