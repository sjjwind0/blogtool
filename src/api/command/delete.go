package command

import (
	"api/info"
	"api/net"
	"errors"
	"fmt"
	"framework"
	"framework/base/config"
	"framework/base/json"
	fn "framework/net"
	"os"
	"path/filepath"
	"strconv"
)

type DeleteCommand struct {
}

func (a *DeleteCommand) CommandName() string {
	return "delete"
}

func (a *DeleteCommand) Run(arguments ...string) (bool, error) {
	if len(arguments) != 2 {
		return false, errors.New("argument length error, please see delete --help to get more info.")
	}
	switch arguments[0] {
	case "local":
		index, err := strconv.Atoi(arguments[1])
		if err != nil {
			return false, nil
		}
		a.deleteLocal(index)
	case "server":
		blogId, err := strconv.Atoi(arguments[1])
		if err != nil {
			return false, nil
		}
		a.deleteLocal(blogId)
	default:
		return false, errors.New("param error")
	}
	return true, nil
}

func (a *DeleteCommand) Usage() string {
	return `delete -u name -p password`
}

func (a *DeleteCommand) deleteLocal(index int) {
	name, sort := info.ShareBlogMapInstance().QueryLocalBlogInfo(index)
	localBlogPath := config.GetDefaultConfigJsonReader().Get("storage.blog").(string)
	localBlogPath = filepath.Join(localBlogPath, sort, name)
	os.RemoveAll(localBlogPath)
}

func (a *DeleteCommand) deleteServer(blogId int) {
	net.StartAPI("/personal/delete", nil, map[string]interface{}{
		"id": blogId,
	}, func(err error, response *fn.HttpResponse) {
		if err == nil && response.Code == 200 {
			reader := json.NewJsonReader(response.Writer.(*fn.StringResponseWriter).GetResponseString())
			c := reader.Get("code").(int64)
			if c == framework.ErrorOK {
				fmt.Println("delete success")
			} else {
				fmt.Println("faield code: ", c)
			}
		} else {
			fmt.Println("err: ", err.Error())
		}
	})
}
