package net

import (
	"fmt"
	"framework/base/config"
	"framework/net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type UploadResponse func(error, *net.HttpResponse)

type UploadProtocol struct {
	net.BaseHttpProtocol
	files    []string
	names    []string
	response UploadResponse
}

func StartUpload(names []string, files []string, response UploadResponse) error {
	defaultConfig := config.GetDefaultConfigJsonReader()
	uploadProtocol := &UploadProtocol{}
	uploadProtocol.Host = defaultConfig.Get("net.host").(string)
	uploadProtocol.Method = "POST"
	uploadProtocol.Path = "/personal/file"
	if defaultConfig.Get("net.protocol").(string) == "http" {
		uploadProtocol.Type = net.HTTP
	} else {
		uploadProtocol.Type = net.HTTPS
	}
	uploadProtocol.Delegate = uploadProtocol
	uploadProtocol.Request = uploadProtocol
	uploadProtocol.Writer = net.NewStringResponseWriter()
	uploadProtocol.response = response
	uploadProtocol.IsSync = true
	uploadProtocol.files = files
	uploadProtocol.names = names
	return uploadProtocol.Start(uploadProtocol)
}

func (u *UploadProtocol) GetHeader() (map[string][]string, []*http.Cookie) {
	return map[string][]string{
		"Connection": []string{"keep-alive"},
		"Accept":     []string{"*/*"},
	}, apiAuthCookies
}

func (u *UploadProtocol) GetURLParams() map[string]string {
	return nil
}

func (u *UploadProtocol) GetBody() []*net.HttpFileInfo {
	var retHttpFileInfoList []*net.HttpFileInfo = nil
	for index, file := range u.files {
		fmt.Println("file: ", file)
		f, err := os.Open(file)
		if err != nil {
			fmt.Println("open file error: ", err.Error())
		}
		defer f.Close()
		fileInfo, err := os.Stat(file)
		if err != nil {
			fmt.Println("open file error: ", err.Error())
		}
		stringContent := make([]byte, fileInfo.Size())
		f.Read(stringContent)
		httpFile := &net.HttpFileInfo{
			Name:        u.names[index],
			FileName:    filepath.Base(file),
			Reader:      strings.NewReader(string(stringContent)),
			ContentSize: fileInfo.Size(),
		}
		retHttpFileInfoList = append(retHttpFileInfoList, httpFile)
	}
	return retHttpFileInfoList
}

func (u *UploadProtocol) OnUploadProgress(current int64, total int64) {

}

func (u *UploadProtocol) OnDownloadProgress(current int64, total int64) {
	fmt.Printf("\bdownload -- %.2f", float64(current)/float64(total))
}

func (u *UploadProtocol) OnComplete(err error, response *net.HttpResponse) {
	if err == nil {
		apiAuthCookies = response.Cookies
		fmt.Println("\bdownload complete")
	}
	u.response(err, response)
}
