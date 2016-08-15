package command

import (
	"api/info"
	"api/net"
	"encoding/json"
	"errors"
	"fmt"
	fn "framework/net"
)

type FetchCommand struct {
}

func (a *FetchCommand) CommandName() string {
	return "fetch"
}

func (a *FetchCommand) Run(arguments ...string) (bool, error) {
	if len(arguments) != 0 {
		return false, errors.New("argument length error, please see fetch --help to get more info.")
	}
	a.fetchAllBlog("blog")
	return true, nil
}

func (a *FetchCommand) Usage() string {
	return `fetch -u name -p password`
}

func (a *FetchCommand) fetchAllBlog(fetchType string) {
	net.StartAPI("/personal/fetch", nil, map[string]interface{}{
		"type": fetchType,
	}, func(err error, response *fn.HttpResponse) {
		if err == nil && response.Code == 200 {
			var d interface{}
			err = json.Unmarshal([]byte(response.Writer.(*fn.StringResponseWriter).GetResponseString()), &d)
			var data map[string]interface{}
			var ok bool
			if data, ok = d.(map[string]interface{}); !ok {
				fmt.Println("transfer failed")
				return
			}
			var code float64
			if code, ok = data["code"].(float64); ok && int(code) == 0 {
				if blogList, ok := data["data"].([]interface{}); ok && blogList != nil {
					info.ShareBlogMapInstance().ClearNetBlogMap()
					for i, b := range blogList {
						if blog, ok := b.(map[string]interface{}); ok {
							blogId := int(blog["id"].(float64))
							blogName := blog["name"].(string)
							blogSort := blog["sort"].(string)
							blogInfo := &info.BlogInfo{Id: blogId, Name: blogName, Sort: blogSort}
							info.ShareBlogMapInstance().AddNetBlog(blogInfo)
							fmt.Printf("%d. id: %d, name: %s\n", i, blogId, blogName)
						}
					}
				}
			} else {
				fmt.Println("faield code: ", code)
			}
		} else {
			fmt.Println("err: ", err.Error())
		}
	})
}
