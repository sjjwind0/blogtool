package net

import (
	"crypto/tls"
	"errors"
	"fmt"
	"framework/base/json"
	"framework/base/timer"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	HTTP  = iota
	HTTPS = iota
	SPDY  = iota
)

const (
	TransferType_Upload   = iota
	TransferType_Download = iota
)

const kTimerOutTime = 10 * time.Second

type HttpProgressDelegate interface {
	OnUploadProgress(current int64, total int64)
	OnDownloadProgress(current int64, total int64)
	OnComplete(err error, response *HttpResponse)
}

type progressReader struct {
	delegate     HttpProgressDelegate
	reader       io.Reader
	total        int64
	sent         int64
	transferType int
}

func (p *progressReader) Read(content []byte) (int, error) {
	n, err := p.reader.Read(content)
	p.sent += int64(n)
	if p.delegate != nil {
		if p.transferType == TransferType_Upload {
			p.delegate.OnUploadProgress(p.sent, p.total)
		} else {
			p.delegate.OnDownloadProgress(p.sent, p.total)
		}
	}
	return n, err
}

func (p *progressReader) Close() error {
	return nil
}

type JsonRequestSeriailizer interface {
	GetBody() map[string]interface{}
}

type FormRequestSerializer interface {
	GetBody() map[string]string
}

type BinaryRequestSerializer interface {
	GetBody() []byte
}

type HttpFileInfo struct {
	Name        string
	FileName    string
	Reader      io.Reader
	ContentSize int64
}

type FileReuqestSerializer interface {
	GetBody() []*HttpFileInfo
}

type HttpRequest interface {
	GetHeader() (map[string][]string, []*http.Cookie)
	GetURLParams() map[string]string
}

type HttpResponse struct {
	Code    int
	Writer  ResponseWriter
	Cookies []*http.Cookie
	Header  http.Header
}

type BaseHttpProtocol struct {
	Type         int
	Host         string
	Path         string
	Ip           string
	Port         int
	Method       string
	IsSync       bool
	Request      HttpRequest
	Delegate     HttpProgressDelegate
	Writer       ResponseWriter
	timeOutTimer *timer.Timer
	stop         bool
}

func NewBaseHttpProtocol() *BaseHttpProtocol {
	return &BaseHttpProtocol{}
}

func (b *BaseHttpProtocol) getUrl() string {
	var urlParams string = ""
	params := b.Request.GetURLParams()
	for k, v := range params {
		if urlParams != "" {
			urlParams += "&"
		}
		urlParams += k + "=" + v
	}

	var urlString string = ""
	if b.Type == HTTP {
		urlString = "http://"
	} else {
		urlString = "https://"
	}
	urlString += b.Host + b.Path + "?" + urlParams
	fmt.Println("url: ", urlString)
	return urlString
}

func (b *BaseHttpProtocol) getContentTypeAndBody(protocol interface{}) ([]string, string, error) {
	var body string = ""
	var contentType []string = nil
	if p, ok := protocol.(JsonRequestSeriailizer); ok {
		contentType = []string{"application/json", "charset=utf-8"}
		body = json.ToJsonString(p.GetBody())
	} else if p, ok := protocol.(FormRequestSerializer); ok {
		contentType = []string{"application/x-www-form-urlencoded; charset=utf-8"}
		bodyMap := p.GetBody()
		for k, v := range bodyMap {
			if body != "" {
				body += "&"
			}
			body += k + "=" + v
		}
	} else if p, ok := protocol.(FileReuqestSerializer); ok {
		boundry := "1d83f6f8fd284b829763c93858b07e7c"
		contentType = []string{"multipart/form-data; boundary=" + boundry}
		files := p.GetBody()
		for _, file := range files {
			body += "--" + boundry + "\r\n"
			body += `Content-Disposition: form-data; name="` + file.Name + `"; filename="` + file.FileName + "\"\r\n"
			body += "Content-Type: application/octet-stream; charset=utf-8\r\n"
			var content []byte = make([]byte, file.ContentSize)
			var currentReadSize int64 = 0
			for {
				s, err := file.Reader.Read(content[currentReadSize:])
				if err != nil {
					fmt.Println("err: ", err.Error())
					return nil, "", errors.New("read file error")
				}
				currentReadSize += int64(s)
				if s == 0 || currentReadSize == file.ContentSize {
					break
				}
			}
			body += "\r\n" + string(content) + "\r\n"
		}
		body += "--" + boundry + "--\r\n"
	} else if p, ok := protocol.(BinaryRequestSerializer); ok {
		contentType = []string{"application/octet-stream; charset=utf-8"}
		body = string(p.GetBody())
	} else {
		// 没有body
		if b.Method == "POST" {
			return nil, "", errors.New("no post data")
		}
	}
	return contentType, body, nil
}

func (b *BaseHttpProtocol) StartTimeOutTimer(delay time.Duration) {
	if b.timeOutTimer != nil {
		b.timeOutTimer.Stop()
	}
	b.timeOutTimer = timer.NewOneShotTimer()
	b.timeOutTimer.Start(delay, func() {
		fmt.Println("protocol timer out")
		// stop net
		b.stop = true
		b.Delegate.OnComplete(errors.New("Time out"), nil)
	})
}

func (b *BaseHttpProtocol) startRequest(httpRequest *http.Request) {
	b.stop = false
	f := func(httpRequest *http.Request, Delegate HttpProgressDelegate) {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		var err error = nil
		response, err := client.Do(httpRequest)
		if err != nil {
			b.timeOutTimer.Stop()
			Delegate.OnComplete(err, nil)
			return
		}
		var kReadSize int = 4096
		var readByteList []byte = make([]byte, kReadSize)
		var readTotalBytes int64 = 0
		for {
			if b.stop {
				tr.CancelRequest(httpRequest)
				b.timeOutTimer.Stop()
				return
			}
			var readSize int = 0
			readSize, err = response.Body.Read(readByteList)
			if err != nil && readSize == 0 {
				if err != io.EOF {
					b.Writer.Complete(false)
				} else {
					err = nil
				}
				break
			}
			err = nil
			readTotalBytes += int64(readSize)
			b.StartTimeOutTimer(kTimerOutTime)
			Delegate.OnDownloadProgress(readTotalBytes, response.ContentLength)
			err = b.Writer.Write(readByteList, readSize)
			if err != nil {
				fmt.Println("write failed: ", err.Error())
				b.Writer.Complete(false)
			}
		}
		b.timeOutTimer.Stop()
		b.Writer.Complete(true)
		Delegate.OnComplete(err, &HttpResponse{
			Code:    response.StatusCode,
			Cookies: response.Cookies(),
			Writer:  b.Writer,
			Header:  response.Header,
		})
	}
	b.StartTimeOutTimer(kTimerOutTime)
	if b.IsSync {
		f(httpRequest, b.Delegate)
	} else {
		go f(httpRequest, b.Delegate)
	}
}

func (b *BaseHttpProtocol) Start(protocol interface{}) error {
	if b.Delegate == nil {
		return errors.New("Delegate is nil")
	}

	// generate url
	u, err := url.Parse(b.getUrl())
	if err != nil {
		return err
	}

	// generate headers
	headers, cookies := b.Request.GetHeader()
	if headers == nil {
		headers = make(http.Header)
	}

	// generate content type and body
	contentType, body, err := b.getContentTypeAndBody(protocol)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// generate body
	if contentType != nil {
		headers["Content-Type"] = contentType
	}

	// generate progress reader
	uploadProgressReader := &progressReader{
		transferType: TransferType_Upload,
		reader:       strings.NewReader(body),
		total:        int64(len(body)),
		sent:         0,
		delegate:     b.Delegate,
	}

	httpRequest := &http.Request{
		Host:          b.Host,
		Method:        b.Method,
		URL:           u,
		Header:        headers,
		ContentLength: int64(len(body)),
		Body:          uploadProgressReader,
	}
	if cookies != nil {
		for _, cookie := range cookies {
			httpRequest.AddCookie(cookie)
		}
	}

	b.startRequest(httpRequest)
	return nil
}
