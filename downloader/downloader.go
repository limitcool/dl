package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

// refer: https://blog.csdn.net/qq_41035588/article/details/123123001
type Downloader struct {
	MaxProcess int
}

func NewDownloader(maxProcess int) *Downloader {
	return &Downloader{
		MaxProcess: maxProcess,
	}
}

func (d *Downloader) Download(strUrl, filename string) error {
	if filename == "" {
		filename = path.Base(strUrl)
	}
	resp, err := http.Head(strUrl)
	if err != nil {
		return err
	}
	// 通过Head请求 Accept-Ranges: bytes 字段，判断是否支持断点续传和分段下载
	if resp.StatusCode == http.StatusOK && resp.Header.Get("Accept-Ranges") == "bytes" {
		return d.MultiDownload([]string{strUrl}, filename, int(resp.ContentLength))
	}

	return d.signleDownload(strUrl, filename)
}

func (d *Downloader) MultiDownload(strUrls []string, filename string, contentLen int) error {
	log.Println("MultiDownload start")
	partSize := contentLen / d.MaxProcess
	if partSize == 0 {
		return nil
	}

	// 创建分块文件存放的目录
	partDir := d.getPartDir(filename)
	err := os.MkdirAll(partDir, 0777)
	defer os.RemoveAll(partDir)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(d.MaxProcess)
	rangeStart := 0
	for i := 0; i < d.MaxProcess; i++ {
		go func(i, rangeStart int) {
			defer wg.Done()
			rangeEnd := rangeStart + partSize
			// 最后一部分,总长度不能超过Content-Length
			if i == (d.MaxProcess - 1) {
				log.Println(i, "rangeEnd:S", rangeStart, rangeEnd, contentLen)
				rangeEnd = contentLen
				log.Println(i, "rangeEnd:E", rangeStart, rangeEnd)
			}
			err := d.downloadPartial(strUrls[0], filename, int64(rangeStart), int64(rangeEnd), int64(i))
			if err != nil {
				log.Println(err)
			}
		}(i, rangeStart)
		// 这里的分区边界每次都需要额外+1
		// e.g. 因为0-2000,2000的位置已经使用过了,所以下一次的起始位置是2001
		rangeStart += partSize + 1
	}
	wg.Wait()
	// log.Panic("下载完成")
	// 合并文件
	d.merge(filename)
	return nil
}

func (d *Downloader) signleDownload(strUrl string, filename string) error {
	log.Println("signleDownload start")
	resp, err := http.Get(strUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	destFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer destFile.Close()
	io.Copy(destFile, resp.Body)
	return nil
}

// 下载分块文件
func (d *Downloader) downloadPartial(strURL string, filename string, RangeStart, RangeEnd, i int64) error {
	if RangeStart >= RangeEnd {
		return nil
	}

	req, err := http.NewRequest("GET", strURL, nil)
	if err != nil {
		return err
	}
	// 请求Header设置Range
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", RangeStart, RangeEnd))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	flags := os.O_CREATE | os.O_WRONLY
	partfile, err := os.OpenFile(d.getPartFilename(filename, i), flags, 0666)
	if err != nil {
		return err
	}
	defer partfile.Close()

	buf := make([]byte, 30*1024)
	//每次最大写入30MB
	_, err = io.CopyBuffer(partfile, resp.Body, buf)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	return nil
}

// getPartDir 分块文件存放的目录
func (d *Downloader) getPartDir(filename string) string {
	// 将文件名拆分为两块,并取最左侧的部分
	return strings.SplitN(filename, ".", 2)[0]
}

// 构造分块文件的文件名
func (d *Downloader) getPartFilename(filename string, partNum int64) string {
	partDir := d.getPartDir(filename)
	return fmt.Sprintf("%s/%s-%d.part", partDir, filename, partNum)
}

// 合并文件
func (d *Downloader) merge(filename string) error {
	destFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer destFile.Close()
	for i := 0; i < d.MaxProcess; i++ {
		partFileName := d.getPartFilename(filename, int64(i))
		partFile, err := os.Open(partFileName)
		if err != nil {
			return err
		}
		io.Copy(destFile, partFile)
		partFile.Close()
		os.Remove(partFileName)
	}
	return nil
}
