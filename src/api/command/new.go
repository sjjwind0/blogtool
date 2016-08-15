package command

import (
	"errors"
	"fmt"
	"framework/base/config"
	"framework/base/json"
	"github.com/satori/go.uuid"
	"os"
	"path/filepath"
)

type NewCommand struct {
}

func (n *NewCommand) CommandName() string {
	return "new"
}

func (n *NewCommand) Run(arguments ...string) (bool, error) {
	if len(arguments) != 4 {
		return false, errors.New("argument length error, please see new --help to get more info.")
	}
	var sort string = ""
	var name string = ""
	for i := 0; i < len(arguments); i++ {
		if arguments[i] == "-s" {
			i++
			sort = arguments[i]
		}
		if arguments[i] == "-n" {
			i++
			name = arguments[i]
		}
	}
	err := n.newBlog(sort, name)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (n *NewCommand) Usage() string {
	return `delete -u name -p password`
}

func (n *NewCommand) newBlog(sort string, name string) error {
	localBlogPath := config.GetDefaultConfigJsonReader().Get("storage.blog").(string)
	localBlogPath = filepath.Join(localBlogPath, sort, name)
	err := os.MkdirAll(localBlogPath, 0755)
	if err != nil {
		return err
	}
	blogUUID := uuid.NewV4().String()
	// new html.md
	blogFilePath := filepath.Join(localBlogPath, name+".md")
	f, err := os.Create(blogFilePath)
	if err != nil {
		return err
	}
	f.Close()
	// new blog.info
	blogInfo := json.ToJsonString(map[string]interface{}{
		"title":    name,
		"sort":     sort,
		"uuid":     blogUUID,
		"tag":      "",
		"descript": "",
	})
	blogInfoFilePath := filepath.Join(localBlogPath, "blog.info")
	blogInfoFile, err := os.Create(blogInfoFilePath)
	if err != nil {
		return err
	}
	blogInfoFile.WriteString(blogInfo)
	blogInfoFile.Close()
	// new all res folder
	folderList := []string{"res", "res/img", "res/css", "res/html", "res/font", "res/other", "res/js", "res/video"}
	for _, folder := range folderList {
		currentPath := filepath.Join(localBlogPath, folder)
		fmt.Println("current: ", currentPath)
		err := os.MkdirAll(currentPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
