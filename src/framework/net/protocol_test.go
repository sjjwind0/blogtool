package net

import (
	"fmt"
	"testing"
)

var globalChan chan bool = make(chan bool)

type testGetProtocol struct {
	BaseHttpProtocol
}

func newTestGetProtocol(host string) *testGetProtocol {
	p := &testGetProtocol{}
	p.Host = host
	p.Method = "GET"
	p.Path = "/"
	p.Type = HTTP
	p.delegate = p
	p.request = p
	return p
}

func (t *testGetProtocol) GetHeader() map[string][]string {
	return nil
}

func (t *testGetProtocol) GetURLParams() map[string]string {
	return nil
}

func (t *testGetProtocol) OnUploadProgress(current int64, total int64) {
	fmt.Printf("upload: %d/%d\n", current, total)
}

func (t *testGetProtocol) OnDownloadProgress(current int64, total int64) {
	fmt.Printf("download: %d/%d\n", current, total)
}

func (t *testGetProtocol) OnComplete(err error, code int, response string) {
	if err == nil {
		fmt.Println("success code: ", code)
	} else {
		fmt.Printf("error: %s\n", err.Error())
	}
	globalChan <- true
}

func Test_GetStatusOK(t *testing.T) {
	fmt.Println("Test_GetStatusOK")
	p := newTestGetProtocol("www.baidu.com")
	p.Start(p)
	<-globalChan
}

func Test_GetTimeOut(t *testing.T) {
	fmt.Println("Test_GetTimeOut")
	p := newTestGetProtocol("www.facebook.com")
	p.Start(p)
	<-globalChan
}

type testPostJsonProtocol struct {
	BaseHttpProtocol
	body map[string]interface{}
}

func newTestJsonProtocol(host, path string, body map[string]interface{}) *testPostJsonProtocol {
	p := &testPostJsonProtocol{}
	p.Host = host
	p.Method = "POST"
	p.Path = path
	p.Type = HTTP
	p.delegate = p
	p.request = p
	p.body = body
	return p
}

func (t *testPostJsonProtocol) GetHeader() map[string][]string {
	return nil
}

func (t *testPostJsonProtocol) GetURLParams() map[string]string {
	return nil
}

func (t *testPostJsonProtocol) GetBody() map[string]interface{} {
	return t.body
}

func (t *testPostJsonProtocol) OnUploadProgress(current int64, total int64) {
	fmt.Printf("upload: %d/%d\n", current, total)
}

func (t *testPostJsonProtocol) OnDownloadProgress(current int64, total int64) {
	fmt.Printf("download: %d/%d\n", current, total)
}

func (t *testPostJsonProtocol) OnComplete(err error, code int, response string) {
	if err == nil {
		fmt.Println("success code: ", code)
	} else {
		fmt.Printf("error: %s\n", err.Error())
	}
	globalChan <- true
}

func Test_PostJsonStatusOK(t *testing.T) {
	fmt.Println("Test_PostJsonStatusOK")
	p := newTestJsonProtocol("www.baidu.com", "/xxx", map[string]interface{}{
		"a": "b",
		"b": "c",
	})
	p.Start(p)
	<-globalChan
}
