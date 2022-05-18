package downloader

import (
	"strings"
	"testing"
)

func TestSignleDownload(t *testing.T) {
	d := NewDownloader(16)
	err := d.signleDownload("http://nginx.org/download/nginx-1.21.6.tar.gz", "NGINX.tar.gz")
	if err != nil {
		t.Error(err)
	}
	t.Log("OK")
}

func TestSplitN(t *testing.T) {
	t.Log("TestSplitN start")
	str := "nginx.tar.gz"
	strs := strings.SplitN(str, ".", 2)[0]
	t.Log(strs)
}
