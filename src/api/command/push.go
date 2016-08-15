package command

import (
	"api/info"
	"api/net"
	"archive/zip"
	"errors"
	"fmt"
	"framework/base/config"
	"framework/base/json"
	fn "framework/net"
	"github.com/satori/go.uuid"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type PushCommand struct {
}

func (p *PushCommand) CommandName() string {
	return "push"
}

func (p *PushCommand) Run(arguments ...string) (bool, error) {
	if len(arguments) != 1 {
		return false, errors.New("argument length error, please see push -- help to get more info.")
	}
	index, err := strconv.Atoi(arguments[0])
	if err != nil {
		return false, err
	}

	p.pushBlog(index)
	return true, nil
}

func (p *PushCommand) Usage() string {
	return `push index`
}

/*
** @version 1
** 本地文件目录格式如下
**	blog:
**		- sort
**			- name
**				- name.html
**				- name.md
**				- name.info
**				- cover.jpg (固定尺寸)
**				- res
**					- html
**		 			- css
**		 			- js
**					- img
**		 			- font
** raw文件放在raw目录不对外开放，html文件以及res文件放在blog目录，meta信息放数据库。
 */

func (p *PushCommand) compressPath(source string, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()
	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if filepath.Base(path) == ".DS_Store" {
			return nil
		}
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func (p *PushCommand) buildRawZip(zipFilePath string, blogPath string) {
	os.Remove(zipFilePath)
	p.compressPath(blogPath, zipFilePath)
}

func (p *PushCommand) buildAtomMarkDownHtmlFile(newPath, htmlPath string, uuid string) error {
	// 1. 把所有的style抽出来，生成atom_article.css，然后引入atom_article.css
	f, err := os.Open(htmlPath)
	if err != nil {
		fmt.Println("open html file error: ", err.Error())
		return err
	}
	defer f.Close()
	fileInfo, err := os.Stat(htmlPath)
	if err != nil {
		fmt.Println("stat html file error: ", err.Error())
		return err
	}
	var byteContent []byte = make([]byte, fileInfo.Size())
	f.Read(byteContent)
	stringContent := string(byteContent)

	// 处理css
	beginStyle := strings.Index(stringContent, "<style>")
	endStyle := strings.Index(stringContent, "</style>") + len("</style>")

	lastPart := stringContent[endStyle:]
	lastStyleBegin := strings.Index(lastPart, `<link rel="stylesheet" href="`)
	lastStyleEnd := strings.Index(lastPart, ">") + 1
	lastPart = lastPart[:lastStyleBegin] + `<link rel="stylesheet" href="/css/katex.min.css">` + lastPart[lastStyleEnd:]

	insertHtml := `<link rel="stylesheet" href="/css/atom-article.css">`
	stringContent = stringContent[:beginStyle] + insertHtml + lastPart

	// 处理img链接，将res/img/1.jpg修改为article/uuid/img/1.jpg
	re, err := regexp.Compile(`<img src=("(?P<first>.+)").*?>`)
	if err != nil {
		fmt.Println("err: ", err.Error())
		return nil
	}
	allImgSrc := re.FindAllString(stringContent, -1)
	for _, imgUrl := range allImgSrc {
		imgUrl = imgUrl[10:]
		fmt.Println("uuid: ", uuid)
		fmt.Println("imgUrl: ", imgUrl)
		if !(strings.HasPrefix(imgUrl, "http://") || strings.HasPrefix(imgUrl, "https://")) {
			newUrl := "article/" + uuid + imgUrl[3:]
			fmt.Println("imgUrl: ", imgUrl)
			fmt.Println("newUrl: ", newUrl)
			stringContent = strings.Replace(stringContent, imgUrl, newUrl, -1)
		}
	}
	// 处理a href链接
	re, err = regexp.Compile(`<a href="(.*)".*?>`)
	if err != nil {
		fmt.Println("err: ", err.Error())
		return nil
	}
	allHrefSrc := re.FindAllString(stringContent, -1)
	for _, hrefUrl := range allHrefSrc {
		hrefUrl = hrefUrl[9:]
		if !(strings.HasPrefix(hrefUrl, "http://") || strings.HasPrefix(hrefUrl, "https://")) {
			newUrl := "article/" + uuid + hrefUrl[3:]
			stringContent = strings.Replace(stringContent, hrefUrl, newUrl, -1)
		}
	}

	os.Remove(newPath)
	newFile, err := os.Create(newPath)
	if err != nil {
		fmt.Println("creat html file error: ", err.Error())
		return err
	}
	defer newFile.Close()
	newFile.Write([]byte(stringContent))
	return nil
}

func (p *PushCommand) handleBlogMetaInfoFile(blogInfoPath string) string {
	blogMetaInfo := json.NewJsonReaderFromFile(blogInfoPath)
	blogUUID := blogMetaInfo.Get("uuid").(string)
	blogDescription := blogMetaInfo.Get("descript").(string)
	blogTag := blogMetaInfo.Get("tag").(string)
	blogTitle := blogMetaInfo.Get("title").(string)

	if blogUUID == "" {
		blogUUID = uuid.NewV4().String()
	}

	writeMetaInfo := map[string]interface{}{
		"title":    blogTitle,
		"descript": blogDescription,
		"tag":      blogTag,
		"uuid":     blogUUID,
	}
	metaInfoString := json.ToJsonString(writeMetaInfo)
	fmt.Println("metaInfoString: ", metaInfoString)
	f, err := os.Open(blogInfoPath)
	if err != nil {
		fmt.Print("handleBlogMetaInfoFile error: ", err.Error())
		return ""
	}
	f.Write([]byte(metaInfoString))
	f.Close()
	return blogUUID
}

func (p *PushCommand) buildResZipFile(resZipFilePath string, resPath string) {
	os.Remove(resZipFilePath)
	p.compressPath(resPath, resZipFilePath)
}

func (p *PushCommand) pushBlog(index int) {
	blogPath := config.GetDefaultConfigJsonReader().Get("storage.blog").(string)

	blogName, sortType := info.ShareBlogMapInstance().QueryLocalBlogInfo(index)

	blogPath = filepath.Join(blogPath, sortType, blogName)

	tmpPath := config.GetDefaultConfigJsonReader().Get("storage.tmp").(string)

	// 1. 处理blog.info
	blogInfoPath := filepath.Join(blogPath, "blog.info")
	uuid := p.handleBlogMetaInfoFile(blogInfoPath)

	// 2. 处理name.html
	htmlPath := filepath.Join(blogPath, blogName+".html")
	handleHtmlPath := filepath.Join(tmpPath, sortType+"_"+blogName+"_generate.html")
	p.buildAtomMarkDownHtmlFile(handleHtmlPath, htmlPath, uuid)

	// 3. 生成raw.zip
	rawZipFilePath := filepath.Join(tmpPath, sortType+"_"+blogName+"_raw.zip")
	p.buildRawZip(rawZipFilePath, blogPath)

	// 4. 压缩res.zip
	resFilePath := filepath.Join(blogPath, "res")
	resZipFilePath := filepath.Join(tmpPath, sortType+"_"+blogName+"_res.zip")
	p.buildResZipFile(resZipFilePath, resFilePath)

	// 5. 组装所有的文件
	coverImgPath := filepath.Join(blogPath, "cover.jpg")
	allFileName := []string{"raw", "web", "info", "res", "img"}
	allFilePath := []string{rawZipFilePath, handleHtmlPath, blogInfoPath, resZipFilePath, coverImgPath}

	net.StartUpload(allFileName, allFilePath, func(err error, response *fn.HttpResponse) {
		if err == nil && response.Code == 200 {
			fmt.Println("download success")
		} else if err != nil {
			fmt.Println("err: ", err.Error())
		} else {
			fmt.Println("http OK: ", response.Code)
		}
	})
}
