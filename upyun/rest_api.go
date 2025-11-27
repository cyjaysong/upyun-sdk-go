package upyun

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"
)

// Upload 上传
// filePath 不包含 Bucket
func (c *Client) Upload(savePath string, content []byte, opts ...UploadOption) error {
	authorization, date, reqPath := c.GetHeaderAuthorization("PUT", savePath, "")

	req := c.req.R().SetHeader("Authorization", authorization).SetHeader("Date", date)
	req.SetBody(content)
	// 应用上传选项
	for _, opt := range opts {
		opt(req)
	}

	resp, err := req.Put(reqPath)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("upload failed: %s", resp.String())
	}
	return nil
}

// Delete 删除
// path 不包含 Bucket
func (c *Client) Delete(path string) error {
	authorization, date, reqPath := c.GetHeaderAuthorization("DELETE", path, "")

	req := c.req.R().SetHeader("Authorization", authorization).SetHeader("Date", date)

	resp, err := req.Delete(reqPath)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("delete failed: %s", resp.String())
	}
	return nil
}

// Copy 复制
// srcPath, destPath 不包含 Bucket
func (c *Client) Copy(srcPath, destPath string) error {
	authorization, date, reqPath := c.GetHeaderAuthorization("PUT", destPath, "")
	srcPath = path.Join(append([]string{"/", c.config.Bucket}, strings.Split(srcPath, "/")...)...)

	req := c.req.R().SetHeader("Authorization", authorization).SetHeader("Date", date).
		SetHeader("X-Upyun-Copy-Source", srcPath)

	resp, err := req.Put(reqPath)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("copy failed: %s", resp.String())
	}
	return nil
}

// Move 移动
// srcPath, destPath 不包含 Bucket
func (c *Client) Move(srcPath, destPath string) error {
	authorization, date, reqPath := c.GetHeaderAuthorization("PUT", destPath, "")
	srcPath = path.Join(append([]string{"/", c.config.Bucket}, strings.Split(srcPath, "/")...)...)

	req := c.req.R().SetHeader("Authorization", authorization).SetHeader("Date", date).
		SetHeader("X-Upyun-Move-Source", srcPath)

	resp, err := req.Put(reqPath)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("move failed: %s", resp.String())
	}
	return nil
}

// Download 下载
// filePath 不包含 Bucket
func (c *Client) Download(filePath string) ([]byte, error) {
	authorization, date, reqPath := c.GetHeaderAuthorization("GET", filePath, "")

	req := c.req.R().SetHeader("Authorization", authorization).SetHeader("Date", date)

	resp, err := req.Get(reqPath)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download failed: %s", resp.String())
	}
	return resp.Bytes(), nil
}

// List 获取目录文件列表
func (c *Client) List(dirPath, iter string, limit int, desc bool) (files []*FileInfo, nextIter string, err error) {
	if limit <= 0 || limit > 10000 {
		limit = 100
	}
	authorization, date, reqPath := c.GetHeaderAuthorization("GET", dirPath, "")
	req := c.req.R().SetHeader("Authorization", authorization).SetHeader("Date", date).
		SetHeader("X-UpYun-Folder", "true").SetHeader("X-List-Limit", strconv.Itoa(limit))
	if desc {
		req.SetHeader("X-List-Order", "desc")
	}
	if iter != "" {
		req.SetHeader("X-List-Iter", iter)
	}
	resp, err := req.Get(reqPath)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("get obj list failed: %s", resp.String())
	}
	//fmt.Println(resp.HeaderToString())
	if nextIter = resp.GetHeader("X-Upyun-List-Iter"); nextIter == "g2gCZAAEbmV4dGQAA2VvZg" {
		nextIter = ""
	}
	if len(resp.Bytes()) == 0 {
		return nil, nextIter, nil
	}
	fileInfoBytes := bytes.Split(resp.Bytes(), []byte("\n"))
	files = make([]*FileInfo, 0, len(fileInfoBytes))
	for _, fileBytes := range fileInfoBytes {
		fileInfo := bytes.Split(fileBytes, []byte("\t"))
		fileItem := &FileInfo{Name: string(fileInfo[0]), IsDir: string(fileInfo[1]) == "F"}
		fileItem.Size, _ = strconv.ParseInt(string(fileInfo[2]), 10, 64)
		unix, _ := strconv.ParseInt(string(fileInfo[3]), 10, 64)
		fileItem.Time = time.Unix(unix, 0)
		files = append(files, fileItem)
	}
	return
}

type IterationObjsConf struct {
	Dir         string
	ObjsChan    chan *FileInfo
	QuitChan    chan struct{}
	EachObjsNum int  // 单次获取对象数量
	MaxObjsNum  int  // 获取最大数量
	MaxDirDepth int  // 最大目录深度
	DescOrder   bool // 是否根据文件名倒序排序

	rDir   string
	depth  int // 当前目录深度
	objNum int // 对象数量

}

// Iteration 迭代目录文件列表
func (c *Client) Iteration(conf *IterationObjsConf) (err error) {
	if conf.ObjsChan == nil {
		return errors.New("ObjsChan is nil")
	}
	if conf.QuitChan == nil {
		conf.QuitChan = make(chan struct{})
	}
	if conf.depth == 0 {
		defer close(conf.ObjsChan)
	}
	var currentDir = path.Join(conf.Dir, conf.rDir)
	var objs []*FileInfo
	var iter string
	for {
		if objs, iter, err = c.List(currentDir, iter, conf.EachObjsNum, conf.DescOrder); err != nil {
			return err
		}
		for _, obj := range objs {
			if conf.rDir != "" {
				obj.Name = path.Join(conf.rDir, obj.Name)
			}
			select {
			case <-conf.QuitChan:
				return nil
			default:
				conf.ObjsChan <- obj
				conf.objNum++
			}
			if conf.MaxObjsNum > 0 && conf.objNum >= conf.MaxObjsNum {
				return nil
			}
			if obj.IsDir && (conf.MaxDirDepth == -1 || conf.depth < conf.MaxDirDepth) {
				rConf := &IterationObjsConf{
					Dir:         conf.Dir,
					ObjsChan:    conf.ObjsChan,
					QuitChan:    conf.QuitChan,
					EachObjsNum: conf.EachObjsNum,
					MaxObjsNum:  conf.MaxObjsNum,
					MaxDirDepth: conf.MaxDirDepth,
					DescOrder:   conf.DescOrder,
					rDir:        obj.Name,
					depth:       conf.depth + 1,
					objNum:      conf.objNum,
				}
				if err = c.Iteration(rConf); err != nil {
					return err
				}
				conf.objNum = rConf.objNum
				if conf.MaxObjsNum > 0 && conf.objNum >= conf.MaxObjsNum {
					return nil
				}
			}
		}
		if iter == "" {
			return nil
		}
	}
}
