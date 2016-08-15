package net

import (
	"framework/base/config"
	"framework/net"
	"net/http"
)

var apiAuthCookies []*http.Cookie

type APIResponse func(error, *net.HttpResponse)

type APIProtocol struct {
	net.BaseHttpProtocol
	params   map[string]string
	postData map[string]interface{}
	response APIResponse
}

func StartAPI(path string, params map[string]string, postData map[string]interface{},
	response APIResponse) error {
	defaultConfig := config.GetDefaultConfigJsonReader()
	apiProtocol := &APIProtocol{}
	apiProtocol.Host = defaultConfig.Get("net.host").(string)
	apiProtocol.Method = "POST"
	apiProtocol.Path = path
	if defaultConfig.Get("net.protocol").(string) == "http" {
		apiProtocol.Type = net.HTTP
	} else {
		apiProtocol.Type = net.HTTPS
	}
	apiProtocol.Delegate = apiProtocol
	apiProtocol.Request = apiProtocol
	apiProtocol.postData = postData
	apiProtocol.params = params
	apiProtocol.postData = postData
	apiProtocol.response = response
	apiProtocol.Writer = net.NewStringResponseWriter()
	apiProtocol.IsSync = true
	return apiProtocol.Start(apiProtocol)
}

func (a *APIProtocol) GetHeader() (map[string][]string, []*http.Cookie) {
	return nil, apiAuthCookies
}

func (a *APIProtocol) GetURLParams() map[string]string {
	return a.params
}

func (a *APIProtocol) GetBody() map[string]interface{} {
	return a.postData
}

func (a *APIProtocol) OnUploadProgress(current int64, total int64) {
}

func (a *APIProtocol) OnDownloadProgress(current int64, total int64) {
}

func (a *APIProtocol) OnComplete(err error, response *net.HttpResponse) {
	apiAuthCookies = response.Cookies
	a.response(err, response)
}
