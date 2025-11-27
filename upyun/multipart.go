package upyun

//// MultiUploadInfo 并行上传信息
//type MultiUploadInfo struct {
//	UUID     string // 任务标识
//	FileSize int64  // 文件大小
//	PartSize int64  // 分块大小
//	PartNum  int    // 分块数量
//}
//
//// InitiateMultiUpload 初始化并行上传
//func (c *Client) InitiateMultiUpload(path string, fileSize int64, partSize int64, opts ...MultiUploadOption) (*MultiUploadInfo, error) {
//	if partSize == 0 {
//		partSize = 1024 * 1024 // 默认1M
//	}
//
//	// 计算分块数量
//	partNum := int((fileSize + partSize - 1) / partSize)
//
//	uri := fmt.Sprintf("/%s/%s", c.config.Bucket, path)
//	auth := c.GetHeaderAuthorization("PUT", uri, "")
//
//	req := c.req.R().
//		SetHeader("Authorization", auth).
//		SetHeader("Date", time.Now().UTC().Format(time.RFC1123)).
//		SetHeader("X-Upyun-Multi-Disorder", "true").
//		SetHeader("X-Upyun-Multi-Stage", "initiate").
//		SetHeader("X-Upyun-Multi-Length", strconv.FormatInt(fileSize, 10)).
//		SetHeader("X-Upyun-Multi-Part-Size", strconv.FormatInt(partSize, 10))
//
//	// 应用上传选项
//	for _, opt := range opts {
//		opt(req)
//	}
//
//	resp, err := req.Put(fmt.Sprintf("%s%s%s", c.getProtocol(), c.config.Domain, uri))
//	if err != nil {
//		return nil, err
//	}
//
//	if resp.StatusCode != 204 {
//		return nil, fmt.Errorf("initiate multi upload failed: %s", resp.String())
//	}
//
//	// 获取任务标识
//	uuid := resp.Header.Get("X-Upyun-Multi-Uuid")
//	if uuid == "" {
//		return nil, fmt.Errorf("failed to get upload uuid")
//	}
//
//	return &MultiUploadInfo{
//		UUID:     uuid,
//		FileSize: fileSize,
//		PartSize: partSize,
//		PartNum:  partNum,
//	}, nil
//}
//
//// MultiUploadOption 并行上传选项函数类型
//type MultiUploadOption func(*req.Request)
//
//// WithMultiContentType 设置并行上传的Content-Type
//func WithMultiContentType(contentType string) MultiUploadOption {
//	return func(req *req.Request) {
//		req.SetHeader("X-Upyun-Multi-Type", contentType)
//	}
//}
//
//// UploadPart 上传分块
//func (c *Client) UploadPart(path string, info *MultiUploadInfo, partID int, partContent []byte) error {
//	uri := fmt.Sprintf("/%s/%s", c.config.Bucket, path)
//	auth := c.GetHeaderAuthorization("PUT", uri, "")
//
//	resp, err := c.req.R().
//		SetHeader("Authorization", auth).
//		SetHeader("Date", time.Now().UTC().Format(time.RFC1123)).
//		SetHeader("X-Upyun-Multi-Disorder", "true").
//		SetHeader("X-Upyun-Multi-Stage", "upload").
//		SetHeader("X-Upyun-Multi-Uuid", info.UUID).
//		SetHeader("X-Upyun-Part-Id", strconv.Itoa(partID)).
//		SetBody(partContent).
//		Put(fmt.Sprintf("%s%s%s", c.getProtocol(), c.config.Domain, uri))
//
//	if err != nil {
//		return err
//	}
//
//	if resp.StatusCode != 204 {
//		return fmt.Errorf("upload part %d failed: %s", partID, resp.String())
//	}
//
//	return nil
//}
//
//// CompleteMultiUpload 完成并行上传
//func (c *Client) CompleteMultiUpload(path string, info *MultiUploadInfo) error {
//	uri := fmt.Sprintf("/%s/%s", c.config.Bucket, path)
//	auth := c.GetHeaderAuthorization("PUT", uri, "")
//
//	resp, err := c.req.R().
//		SetHeader("Authorization", auth).
//		SetHeader("Date", time.Now().UTC().Format(time.RFC1123)).
//		SetHeader("X-Upyun-Multi-Disorder", "true").
//		SetHeader("X-Upyun-Multi-Stage", "complete").
//		SetHeader("X-Upyun-Multi-Uuid", info.UUID).
//		Put(fmt.Sprintf("%s%s%s", c.getProtocol(), c.config.Domain, uri))
//
//	if err != nil {
//		return err
//	}
//
//	if resp.StatusCode != 200 {
//		return fmt.Errorf("complete multi upload failed: %s", resp.String())
//	}
//
//	return nil
//}
//
//// UploadLargeFile 上传大文件（自动分块）
//func (c *Client) UploadLargeFile(path string, filePath string, partSize int64, opts ...MultiUploadOption) error {
//	// 打开文件
//	file, err := os.Open(filePath)
//	if err != nil {
//		return err
//	}
//	defer file.Close()
//
//	// 获取文件大小
//	fileInfo, err := file.Stat()
//	if err != nil {
//		return err
//	}
//	fileSize := fileInfo.Size()
//
//	// 初始化上传
//	info, err := c.InitiateMultiUpload(path, fileSize, partSize, opts...)
//	if err != nil {
//		return err
//	}
//
//	// 上传所有分块
//	for i := 0; i < info.PartNum; i++ {
//		// 计算分块偏移量和大小
//		offset := int64(i) * info.PartSize
//		currentPartSize := info.PartSize
//		if offset+currentPartSize > fileSize {
//			currentPartSize = fileSize - offset
//		}
//
//		// 读取分块内容
//		partContent := make([]byte, currentPartSize)
//		_, err := file.ReadAt(partContent, offset)
//		if err != nil && err != io.EOF {
//			return err
//		}
//
//		// 上传分块
//		err = c.UploadPart(path, info, i, partContent)
//		if err != nil {
//			return err
//		}
//	}
//
//	// 完成上传
//	return c.CompleteMultiUpload(path, info)
//}
