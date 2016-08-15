package command

import (
	"api/info"
	"api/net"
	"errors"
	"fmt"
	"framework/base/config"
	fn "framework/net"
	"framework/util/archive"
	"os"
	"path/filepath"
	"strconv"
)

type PullCommand struct {
}

func (p *PullCommand) CommandName() string {
	return "pull"
}

func (p *PullCommand) Run(arguments ...string) (bool, error) {
	if len(arguments) != 1 {
		return false, errors.New("argument length error, please see pull -- help to get more info.")
	}
	blogId, err := strconv.Atoi(arguments[0])
	if err != nil {
		return false, err
	}
	p.pullNetworkBlog(blogId)
	return true, nil
}

func (p *PullCommand) Usage() string {
	return `pull blogId`
}

func (p *PullCommand) pullNetworkBlog(blogId int) {
	downloadTmpPath := "/tmp/download_raw.zip"
	net.StartDownload(map[string]string{"id": strconv.Itoa(blogId)}, downloadTmpPath,
		func(err error, response *fn.HttpResponse) {
			if err == nil && response.Code == 200 {
				// unzip zip
				_, sort := info.ShareBlogMapInstance().QueryNetBlogInfo(blogId)
				blogPath := config.GetDefaultConfigJsonReader().Get("storage.blog").(string)
				zipPath := filepath.Join(blogPath, sort, "download_raw.zip")
				err = os.Rename(downloadTmpPath, zipPath)
				if err != nil {
					fmt.Println("os.Rename error: ", err.Error())
				}
				err = archive.UnZip(zipPath)
				os.Remove(zipPath)
				if err != nil {
					fmt.Println("unzip error: ", err.Error())
				}
				fmt.Println("pull success")
			} else {
				fmt.Println("err: ", err.Error())
			}
		})
}
