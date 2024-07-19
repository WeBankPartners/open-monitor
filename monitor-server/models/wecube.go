package models

import (
	"bytes"
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

func refreshToken() error {
	req, err := http.NewRequest(http.MethodGet, CoreUrl+"/auth/v1/api/token", strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("http new request fail,%s ", err.Error())
	}
	req.Header.Set("Authorization", coreRefreshToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http response fail,%s ", err.Error())
	}
	var respObj responseObj
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	err = json.Unmarshal(bodyBytes, &respObj)
	if err != nil {
		return fmt.Errorf("http response body json unmarshal fail,%s ", err.Error())
	}
	for _, v := range respObj.Data {
		if len(v.Expiration) > 10 {
			v.Expiration = v.Expiration[:10]
		}
		expInt, _ := strconv.ParseInt(v.Expiration, 10, 64)
		if v.TokenType == "refreshToken" {
			coreRefreshToken = "Bearer " + v.Token
			coreRefreshTokenExpTime = time.Unix(expInt, 0)
		}
		if v.TokenType == "accessToken" {
			coreRequestToken = "Bearer " + v.Token
			coreRequestTokenExpTime = time.Unix(expInt, 0)
		}
	}
	return nil
}

func requestCoreToken(rsaKey string) error {
	encryptBytes, err := cipher.RSAEncryptByPrivate([]byte(fmt.Sprintf("%s:%s", SubSystemCode, requestCoreNonce)), []byte(rsaKey))
	encryptString := base64.StdEncoding.EncodeToString(encryptBytes)
	if err != nil {
		return err
	}
	postParam := requestToken{Username: SubSystemCode, Nonce: requestCoreNonce, ClientType: "SUB_SYSTEM", Password: encryptString}
	postBytes, _ := json.Marshal(postParam)
	fmt.Printf("param: %s \n", string(postBytes))
	req, err := http.NewRequest(http.MethodPost, CoreUrl+"/auth/v1/api/login", bytes.NewReader(postBytes))
	if err != nil {
		return fmt.Errorf("http new request fail,%s ", err.Error())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http response fail, %s ", err.Error())
	}
	var respObj responseObj
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	err = json.Unmarshal(bodyBytes, &respObj)
	if err != nil {
		return fmt.Errorf("http response body read fail,%s ", err.Error())
	}
	for _, v := range respObj.Data {
		if len(v.Expiration) > 10 {
			v.Expiration = v.Expiration[:10]
		}
		expInt, _ := strconv.ParseInt(v.Expiration, 10, 64)
		if v.TokenType == "refreshToken" {
			coreRefreshToken = "Bearer " + v.Token
			coreRefreshTokenExpTime = time.Unix(expInt, 0)
		}
		if v.TokenType == "accessToken" {
			coreRequestToken = "Bearer " + v.Token
			coreRequestTokenExpTime = time.Unix(expInt, 0)
		}
	}
	return nil
}

func InitCoreToken() {
	err := requestCoreToken(SubSystemKey)
	if err != nil {
		log.Printf("Init core token fail,error: %s ", err.Error())
	} else {
		log.Println("Init core token success")
	}
}

func GetCoreToken() string {
	if CoreUrl == "" {
		return ""
	}
	if coreRequestTokenExpTime.Unix() > time.Now().Unix() && coreRequestToken != "" {
		return coreRequestToken
	}
	if coreRefreshTokenExpTime.Unix() > time.Now().Unix() && coreRefreshToken != "" {
		err := refreshToken()
		if err != nil {
			log.Printf("Refresh token fail,%s ", err.Error())
		} else {
			return coreRequestToken
		}
	}
	err := requestCoreToken(SubSystemKey)
	if err != nil {
		log.Printf("Try to init core token fail,%s ", err.Error())
	}
	return coreRefreshToken
}
