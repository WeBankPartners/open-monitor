package models

import (
	"strings"
	"encoding/base64"
	"io/ioutil"
	"encoding/pem"
	"crypto/x509"
	"crypto/rsa"
	"crypto/rand"
	"log"
)

func DecryptRsa(inputString string) string {
	if !strings.HasPrefix(strings.ToLower(inputString), "rsa@") {
		return inputString
	}
	inputString = inputString[4:]
	result := inputString
	inputBytes,err := base64.RawStdEncoding.DecodeString(inputString)
	if err != nil {
		log.Printf("Input string format to base64 fail,%s \n", err.Error())
		return inputString
	}
	pemPath := "/data/certs/rsa_key"
	fileContent,err := ioutil.ReadFile(pemPath)
	if err != nil {
		log.Printf("Read file %s fail,%s \n", pemPath, err.Error())
		return result
	}
	block,_ := pem.Decode(fileContent)
	privateKeyInterface,err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Printf("Parse private key fail,%s \n", err.Error())
		return result
	}
	privateKey := privateKeyInterface.(*rsa.PrivateKey)
	decodeBytes,err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, inputBytes)
	if err != nil {
		log.Printf("Decode fail,%s \n", err.Error())
		return result
	}
	result = string(decodeBytes)
	return result
}
