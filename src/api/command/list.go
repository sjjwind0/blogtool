package command

import (
	"api/info"
	"errors"
	"fmt"
	"framework/base/config"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type ListCommand struct {
}

func (l *ListCommand) CommandName() string {
	return "list"
}

func (l *ListCommand) Run(arguments ...string) (bool, error) {
	if len(arguments) != 0 {
		return false, errors.New("argument length error, please see login --help to get more info.")
	}
	l.listAllBlog()
	return true, nil
}

func (l *ListCommand) Usage() string {
	return `list -u name -p password`
}

func (l *ListCommand) listAllBlog() {
	blogPath := config.GetDefaultConfigJsonReader().Get("storage.blog").(string)
	files, err := ioutil.ReadDir(blogPath)
	if err != nil {
		fmt.Print("error: ", err.Error())
		return
	}
	info.ShareBlogMapInstance().ClearLocalBlogMap()
	var index int = 1
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		fmt.Println("\x1B[31m" + file.Name())
		if file.IsDir() {
			childs, err := ioutil.ReadDir(filepath.Join(blogPath, file.Name()))
			if err != nil {
				fmt.Print("error: ", err.Error())
				return
			}
			for _, child := range childs {
				if strings.HasPrefix(child.Name(), ".") {
					continue
				}
				fmt.Printf("\x1B[34m%d. %s\n", index, child.Name())
				blogInfo := &info.BlogInfo{Id: index, Name: child.Name(), Sort: file.Name()}
				info.ShareBlogMapInstance().AddLocalBlog(blogInfo)
				index++
			}
		}
	}
}
