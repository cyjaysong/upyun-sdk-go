package upyun

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	reqUtil "github.com/imroc/req/v3"
)

// Config 配置结构体
type Config struct {
	passwordMD5Bytes []byte
	baseUrl          string

	Bucket    string
	Operator  string
	Password  string
	Domain    string // 可选，默认使用智能选路域名 v0.api.upyun.com
	UserAgent string // 可选，用于设置请求时的UserAgent
	UseHTTP   bool   // 可选，是否使用HTTP协议，默认使用HTTPS
}

// Client 又拍云客户端
type Client struct {
	config *Config
	req    *reqUtil.Client
}

// NewClient 创建新的又拍云客户端
func NewClient(config *Config) *Client {
	client := &Client{config: config, req: reqUtil.C()}
	client.config.passwordMD5Bytes = []byte(strMd5(config.Password))
	// 设置默认Domain
	if config.Domain == "" {
		client.config.Domain = "v0.api.upyun.com"
	}
	// 生成BaseUrl 如 https://v0.api.upyun.com
	client.config.baseUrl = client.getBaseURL()

	// 设置请求的BaseURL 如https://v0.api.upyun.com
	client.req.SetBaseURL(client.config.baseUrl)
	// 设置UserAgent
	if config.UserAgent != "" {
		client.req.SetUserAgent(config.UserAgent)
	}
	return client
}

func base64ToStr(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func strMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func makeRFC1123Date(d time.Time) string {
	utc := d.UTC().Format(time.RFC1123)
	return strings.Replace(utc, "UTC", "GMT", -1)
}

func (c *Client) getBaseURL() string {
	protocol := "https"
	if c.config.UseHTTP {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s", protocol, c.config.Domain)
}

// GetBaseURL 获取请求BaseURL
func (c *Client) GetBaseURL() string {
	return c.config.baseUrl
}

// VerifyHeaderAuthorization 验证Header认证标识
// path 是用户接口的path
func (c *Client) VerifyHeaderAuthorization(authorization, date, method, path, fileMd5 string) bool {
	mac := hmac.New(sha1.New, c.config.passwordMD5Bytes)
	elements := []string{method, path, date}
	if len(fileMd5) > 0 {
		elements = append(elements, fileMd5)
	}
	value := strings.Join(elements, "&")
	mac.Write([]byte(value))
	signature := base64ToStr(mac.Sum(nil))
	return authorization == fmt.Sprintf("UPYUN %s:%s", c.config.Operator, signature)
}

// GetHeaderAuthorization 获取Header认证标识
// filePath 不包含 Bucket
func (c *Client) GetHeaderAuthorization(method, filePath, fileMd5 string) (authorization, date, reqPath string) {
	reqPath = path.Join("/", c.config.Bucket, filePath)
	if filePath == "/" {
		reqPath += "/"
	}
	timeNow := time.Now()
	date = makeRFC1123Date(timeNow)
	mac := hmac.New(sha1.New, c.config.passwordMD5Bytes)
	elements := []string{method, reqPath, date}
	if len(fileMd5) > 0 {
		elements = append(elements, fileMd5)
	}
	value := strings.Join(elements, "&")
	mac.Write([]byte(value))
	signature := base64ToStr(mac.Sum(nil))
	authorization = fmt.Sprintf("UPYUN %s:%s", c.config.Operator, signature)
	return
}

// GetBodyAuthorization 获取Body认证标识
// filePath 不包含 Bucket
func (c *Client) GetBodyAuthorization(method, filePath, fileMd5 string,
	policyParams map[string]string, expire time.Duration) (authorization, policy, reqPath string) {
	reqPath = path.Join("/", c.config.Bucket)
	timeNow := time.Now()
	date := makeRFC1123Date(timeNow)
	if expire == 0 {
		expire = time.Minute * 30
	}
	mac := hmac.New(sha1.New, c.config.passwordMD5Bytes)
	elements := []string{method, reqPath, date}
	if len(policyParams) == 0 {
		policyParams = make(map[string]string)
	}
	policyParams["bucket"] = c.config.Bucket
	policyParams["expiration"] = strconv.FormatInt(timeNow.Add(expire).Unix(), 10)
	policyParams["save-key"] = filePath
	policyParams["date"] = date
	policyBytes, _ := json.Marshal(policyParams)
	policy = base64ToStr(policyBytes)
	elements = append(elements, policy)
	if len(fileMd5) > 0 {
		elements = append(elements, fileMd5)
	}
	value := strings.Join(elements, "&")
	mac.Write([]byte(value))
	signature := base64ToStr(mac.Sum(nil))
	authorization = fmt.Sprintf("UPYUN %s:%s", c.config.Operator, signature)
	return
}
