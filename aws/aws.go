package aws

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Region string

const (
	USEast1Region Region = "us-east-1"
	USWest2Region        = "us-west-2"
	EUWest1Region        = "eu-west-1"
)

var UnsupportedRegionErr = errors.New("Unsupported region")

type Parameters map[string]interface{}

type ResponseMetadata struct {
	RequestId string `xml:"RequestId"`
}

type Credentials struct {
	accessKeyId string
	secretKey   string
}

func NewCredentials(accessKeyId, secretKey string) Credentials {
	return Credentials{accessKeyId: accessKeyId, secretKey: secretKey}
}

func createRequestUrl(endpoint, path string) (string, error) {
	return endpoint + path, nil
}

func generateSignature(key, message string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func createAuthorizationHeaders(credentials Credentials) (map[string]string, error) {
	date := time.Now().Format(time.RFC1123)
	signature := generateSignature(credentials.secretKey, date)
	return map[string]string{
		"Date":                 date,
		"X-Amzn-Authorization": fmt.Sprintf("AWS3-HTTPS AWSAccessKeyId=%s, Algorithm=HmacSHA256, Signature=%s", credentials.accessKeyId, signature),
	}, nil
}

func createRequestValues(credentials Credentials, action string, parameters Parameters) (url.Values, error) {
	values := url.Values{}
	values.Set("Action", action)
	for key, value := range parameters {
		switch t := value.(type) {
		case string:
			values.Set(key, t)
		case int:
			values.Set(key, strconv.Itoa(t))
		case []string:
			for i, v := range t {
				values.Set(fmt.Sprintf("%s.%d", key, i+1), v)
			}
		default:
			return nil, fmt.Errorf("Unknown type %T for parameter %s", t, key)
		}
	}
	return values, nil
}

func ExecuteRequest(credentials Credentials, endpoint, path, action string, parameters Parameters) ([]byte, error) {
	url, err := createRequestUrl(endpoint, path)
	if err != nil {
		return nil, err
	}
	//log.Println("url: ", url)

	values, err := createRequestValues(credentials, action, parameters)
	if err != nil {
		return nil, err
	}
	//log.Println("values: ", values)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	headers, err := createAuthorizationHeaders(credentials)
	if err != nil {
		return nil, err
	}
	//log.Println("headers: ", headers)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	//log.Println(string(body))

	return body, nil
}
