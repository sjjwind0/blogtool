package net

import (
	"fmt"
	"framework/base/config"
	"framework/net"
	"net/http"
)

type DownloadResponse func(error, *net.HttpResponse)

type DownloadProtocol struct {
	net.BaseHttpProtocol
	params   map[string]string
	response DownloadResponse
}

func StartDownload(param map[string]string, downloadPath string, response DownloadResponse) error {
	defaultConfig := config.GetDefaultConfigJsonReader()
	downloadProtocol := &DownloadProtocol{}
	downloadProtocol.Host = defaultConfig.Get("net.host").(string)
	downloadProtocol.Method = "GET"
	downloadProtocol.Path = "/personal/file"
	if defaultConfig.Get("net.protocol").(string) == "http" {
		downloadProtocol.Type = net.HTTP
	} else {
		downloadProtocol.Type = net.HTTPS
	}
	downloadProtocol.Delegate = downloadProtocol
	downloadProtocol.Request = downloadProtocol
	downloadProtocol.Writer = net.NewFileResponseWriter(downloadPath)
	downloadProtocol.response = response
	downloadProtocol.IsSync = true
	downloadProtocol.params = param
	return downloadProtocol.Start(downloadProtocol)
}

func (d *DownloadProtocol) GetHeader() (map[string][]string, []*http.Cookie) {
	return nil, apiAuthCookies
}

func (d *DownloadProtocol) GetURLParams() map[string]string {
	return d.params
}

func (d *DownloadProtocol) OnUploadProgress(current int64, total int64) {
}

func (d *DownloadProtocol) OnDownloadProgress(current int64, total int64) {
}

func (d *DownloadProtocol) OnComplete(err error, response *net.HttpResponse) {
	if err == nil {
		fmt.Println("\bdownload complete")
	}
	apiAuthCookies = response.Cookies
	d.response(err, response)
}
