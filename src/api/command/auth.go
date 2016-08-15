package command

import (
	"api/net"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"framework"
	"framework/base/json"
	fn "framework/net"
)

type AuthCommand struct {
}

func (a *AuthCommand) CommandName() string {
	return "login"
}

func (a *AuthCommand) Run(arguments ...string) (bool, error) {
	if len(arguments) != 4 {
		return false, errors.New("argument length error, please see login --help to get more info.")
	}
	var userName string = ""
	var password string = ""
	for i := 0; i < len(arguments); i++ {
		if arguments[i] == "-u" {
			i++
			userName = arguments[i]
		}
		if arguments[i] == "-p" {
			i++
			password = arguments[i]
		}
	}
	a.authByUserNameAndPassword(userName, password)
	return true, nil
}

func (a *AuthCommand) Usage() string {
	return `login -u name -p password`
}

func (a *AuthCommand) authByUserNameAndPassword(userName string, password string) {
	sign := func(password string) string {
		md5Ctx := md5.New()
		md5Ctx.Write([]byte(password))
		cipherStr := md5Ctx.Sum(nil)
		return hex.EncodeToString(cipherStr)
	}

	net.StartAPI("/personal/auth", nil, map[string]interface{}{
		"username": userName,
		"password": sign(userName + password),
	}, func(err error, response *fn.HttpResponse) {
		if err == nil && response.Code == 200 {
			reader := json.NewJsonReader(response.Writer.(*fn.StringResponseWriter).GetResponseString())
			c := reader.Get("code").(int64)
			if c == framework.ErrorOK {
				fmt.Println("login success")
			} else {
				fmt.Println("faield code: ", c)
			}
		} else {
			fmt.Println("err: ", err.Error())
		}
	})
}
