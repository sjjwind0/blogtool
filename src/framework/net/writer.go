package net

import (
	"fmt"
	"os"
)

type ResponseWriter interface {
	Write([]byte, int) error
	Complete(success bool)
}

type StringResponseWriter struct {
	byteContent []byte
	strContent  string
}

func NewStringResponseWriter() *StringResponseWriter {
	return &StringResponseWriter{}
}

func (s *StringResponseWriter) Write(buf []byte, size int) error {
	for i := 0; i < size; i++ {
		s.byteContent = append(s.byteContent, buf[i])
	}
	return nil
}

func (s *StringResponseWriter) GetResponseString() string {
	return s.strContent
}

func (s *StringResponseWriter) Complete(success bool) {
	if success {
		s.strContent = string(s.byteContent)
		s.byteContent = nil
	}
}

type FileResponseWriter struct {
	file *os.File
}

func NewFileResponseWriter(filePath string) *FileResponseWriter {
	responseWriter := &FileResponseWriter{}
	_, err := os.Stat(filePath)
	if !os.IsExist(err) {
		responseWriter.file, err = os.Create(filePath)
		if err != nil {
			fmt.Println("create file error: ", err.Error())
		}
	} else {
		responseWriter.file, err = os.Open(filePath)
	}
	return responseWriter
}

func (f *FileResponseWriter) Write(buf []byte, size int) error {
	fmt.Println("write data: ", size)
	_, err := f.file.Write(buf[:size])
	return err
}

func (f *FileResponseWriter) Complete(success bool) {
	if success {
		f.file.Close()
	}
}
