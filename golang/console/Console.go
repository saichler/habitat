package main

import (
	"bufio"
	"fmt"
	"github.com/saichler/habitat/golang/console/commands"
	"github.com/saichler/habitat/golang/console/model"
	"github.com/saichler/habitat/golang/service"
	. "github.com/saichler/utils/golang"
	"os"
	"strings"
)

func main() {
	serviceManager, err := service.NewServiceManager()
	if err != nil {
		Error("Failed to load habitat", err)
		return
	}
	console := &model.Console{}
	console.ServiceManager = serviceManager
	console.Commands = make(map[string]model.Command)
	console.Running = true

	registerCommand(console, &commands.Connect{})
	registerCommand(console, &commands.Help{})

	prompt := "Habitat-" + serviceManager.HID().String() + ">"

	args := os.Args[1:]

	if len(args) == 0 {
		reader := bufio.NewReader(os.Stdin)
		for ; console.Running; {
			fmt.Print(prompt)
			cmd, _ := reader.ReadString('\n')
			cmd = cmd[0 : len(cmd)-1]
			//cmd:=cwcli.readInput()
			handleCommand(console, cmd)
		}
		fmt.Println("Goodbye!")
	} else {
		for _, cmd := range args {
			handleCommand(console, cmd)
		}
	}
}

func registerCommand(console *model.Console, command model.Command) {
	console.Commands[command.Name()] = command
}

func handleCommand(console *model.Console, cmd string) {
	if cmd == "exit" || cmd == "quit" {
		console.Running = false
		return
	} else if cmd == "" {
		return
	}

	args := strings.Split(cmd, " ")
	if args == nil || len(args) == 0 {
		return
	}

	h := console.Commands[args[0]]

	if h == nil {
		fmt.Println("Error! Unknown Command:" + args[0])
		return
	}
	h.Run(console, args)
}
