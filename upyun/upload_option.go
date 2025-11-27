package upyun

import (
	"fmt"

	reqUtil "github.com/imroc/req/v3"
)

// UploadOption 上传选项函数类型
type UploadOption func(*reqUtil.Request)

// WithContentType 设置Content-Type
func WithContentType(contentType string) UploadOption {
	return func(req *reqUtil.Request) {
		req.SetHeader("Content-Type", contentType)
	}
}

// WithContentSecret 设置Content-Secret
func WithContentSecret(secret string) UploadOption {
	return func(req *reqUtil.Request) {
		req.SetHeader("Content-Secret", secret)
	}
}

// WithMetadata 设置元信息
// key不包含 x-upyun-meta-
func WithMetadata(key, value string) UploadOption {
	return func(req *reqUtil.Request) {
		req.SetHeader(fmt.Sprintf("x-upyun-meta-%s", key), value)
	}
}

// WithTTL 设置文件生存时间（天）
func WithTTL(ttl int) UploadOption {
	return func(req *reqUtil.Request) {
		req.SetHeader("x-upyun-meta-ttl", fmt.Sprintf("%d", ttl))
	}
}

// WithGmkerlThumb 设置图片预处理参数
func WithGmkerlThumb(param string) UploadOption {
	return func(req *reqUtil.Request) {
		req.SetHeader("x-gmkerl-thumb", param)
	}
}
