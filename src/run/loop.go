package run

import (
	"api"
	"api/command"
	"bufio"
	"fmt"
	"framework/base/flag"
	"os"
)

type RunLoop struct {
	commandMap map[string]api.Command
}

func (r *RunLoop) registerAPI() {
	if r.commandMap == nil {
		r.commandMap = make(map[string]api.Command)
	}
	addCommand := func(c api.Command) {
		r.commandMap[c.CommandName()] = c
	}
	addCommand(&command.AuthCommand{})
	addCommand(&command.FetchCommand{})
	addCommand(&command.PullCommand{})
	addCommand(&command.ListCommand{})
	addCommand(&command.PushCommand{})
	addCommand(&command.DeleteCommand{})
	addCommand(&command.NewCommand{})
}

func (r *RunLoop) Start() {
	r.registerAPI()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\x1B[0m>>> ")
		line, err := reader.ReadString('\n')
		line = line[:len(line)-1]
		if err != nil {
			fmt.Println("read error: ", err.Error())
			continue
		}
		command, err := flag.Parse(line)
		if err != nil {
			fmt.Println("parse error: ", err.Error())
			continue
		}
		if c, ok := r.commandMap[command.Command]; ok {
			_, err := c.Run(command.Args...)
			if err != nil {
				fmt.Println("command error: ", err.Error())
			}
		} else {
			fmt.Println("unrecogize command, see help for more help.")
		}
	}
}
