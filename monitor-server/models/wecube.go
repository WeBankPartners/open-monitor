package models

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/cipher"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	coreRefreshToken        string
	coreRefreshTokenExpTime time.Time
	coreRequestToken        string
	coreRequestTokenExpTime time.Time
	requestCoreNonce        = "monitor"
	coreTokenHTTPClient     = &http.Client{Timeout: 10 * time.Second}
	coreTokenGate           = make(chan struct{}, 1)
)

type requestToken struct {
	Password   string `json:"password"`
	Username   string `json:"username"`
	Nonce      string `json:"nonce"`
	ClientType string `json:"clientType"`
}

type responseObj struct {
	Status  string             `json:"status"`
	Message string             `json:"message"`
	Data    []*responseDataObj `json:"data"`
}

type responseDataObj struct {
	Expiration string `json:"expiration"`
	Token      string `json:"token"`
	TokenType  string `json:"tokenType"`
}

func refreshToken(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, CoreUrl+"/auth/v1/api/token", strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("http new request fail,%s ", err.Error())
	}
	req.Header.Set("Authorization", coreRefreshToken)
	resp, err := coreTokenHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("http response fail: %w", err)
	}
	return cacheCoreTokenResponse(resp)
}

func requestCoreToken(ctx context.Context, rsaKey string) error {
	encryptBytes, err := cipher.RSAEncryptByPrivate([]byte(fmt.Sprintf("%s:%s", SubSystemCode, requestCoreNonce)), []byte(rsaKey))
	encryptString := base64.StdEncoding.EncodeToString(encryptBytes)
	if err != nil {
		return err
	}
	postParam := requestToken{Username: SubSystemCode, Nonce: requestCoreNonce, ClientType: "SUB_SYSTEM", Password: encryptString}
	postBytes, _ := json.Marshal(postParam)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, CoreUrl+"/auth/v1/api/login", bytes.NewReader(postBytes))
	if err != nil {
		return fmt.Errorf("http new request fail,%s ", err.Error())
	}
	resp, err := coreTokenHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("http response fail: %w", err)
	}
	return cacheCoreTokenResponse(resp)
}

func cacheCoreTokenResponse(resp *http.Response) error {
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("core token HTTP status %d", resp.StatusCode)
	}
	var respObj responseObj
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("http response body read fail: %w", err)
	}
	err = json.Unmarshal(bodyBytes, &respObj)
	if err != nil {
		return fmt.Errorf("http response body json unmarshal fail: %w", err)
	}
	if respObj.Status != "OK" {
		return fmt.Errorf("core token request rejected")
	}
	var access string
	var accessExpiry time.Time
	refresh, refreshExpiry := coreRefreshToken, coreRefreshTokenExpTime
	for _, v := range respObj.Data {
		if v == nil {
			continue
		}
		expInt, parseErr := strconv.ParseInt(v.Expiration, 10, 64)
		if v.TokenType == "accessToken" && parseErr != nil {
			return fmt.Errorf("core token response has no valid access token")
		}
		if parseErr == nil && len(v.Expiration) > 10 {
			expInt, _ = strconv.ParseInt(v.Expiration[:10], 10, 64)
		}
		if v.TokenType == "refreshToken" {
			refresh = "Bearer " + v.Token
			refreshExpiry = time.Unix(expInt, 0)
		}
		if v.TokenType == "accessToken" {
			if strings.TrimSpace(v.Token) == "" {
				return fmt.Errorf("core token response has no valid access token")
			}
			access = "Bearer " + v.Token
			accessExpiry = time.Unix(expInt, 0)
		}
	}
	if access == "" || !accessExpiry.After(time.Now()) {
		return fmt.Errorf("core token response has no valid access token")
	}
	coreRequestToken, coreRequestTokenExpTime = access, accessExpiry
	coreRefreshToken, coreRefreshTokenExpTime = refresh, refreshExpiry
	return nil
}

func InitCoreToken() {
	coreTokenGate <- struct{}{}
	defer func() { <-coreTokenGate }()
	err := requestCoreToken(context.Background(), SubSystemKey)
	if err != nil {
		log.Printf("Init core token fail,error: %s ", err.Error())
	} else {
		log.Println("Init core token success")
	}
}

func GetCoreToken() string {
	token, err := GetCoreTokenContext(context.Background())
	if err != nil {
		log.Printf("Get core token fail,%s ", err.Error())
	}
	return token
}

// GetCoreTokenContext serializes access to the token cache while allowing
// callers to cancel both waiting for another refresh and the HTTP request.
func GetCoreTokenContext(ctx context.Context) (string, error) {
	select {
	case coreTokenGate <- struct{}{}:
		defer func() { <-coreTokenGate }()
	case <-ctx.Done():
		return "", ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if CoreUrl == "" {
		return "", nil
	}
	if coreRequestTokenExpTime.Unix() > time.Now().Unix() && coreRequestToken != "" {
		return coreRequestToken, nil
	}
	if coreRefreshTokenExpTime.Unix() > time.Now().Unix() && coreRefreshToken != "" {
		err := refreshToken(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			log.Printf("Refresh token fail,%s ", err.Error())
		} else {
			return coreRequestToken, nil
		}
	}
	if err := requestCoreToken(ctx, SubSystemKey); err != nil {
		return "", err
	}
	return coreRequestToken, nil
}
