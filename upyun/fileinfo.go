package upyun

import (
	"time"
)

type FileInfo struct {
	Name        string
	Size        int64
	ContentType string
	IsDir       bool
	MD5         string
	Time        time.Time

	Meta map[string]string

	/* image information */
	ImgType   string
	ImgWidth  int64
	ImgHeight int64
	ImgFrames int64
}
