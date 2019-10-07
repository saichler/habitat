package commands

import (
	"bytes"
	"fmt"
	"github.com/saichler/habitat/golang/console/model"
)

type Help struct{}

func (cmd *Help) Name() string {
	return "?"
}

func (cmd *Help) Usage() string {
	return "?"
}

func (cmd *Help) Description() string {
	return "Prints this help."
}

func (cmd *Help) Run(console *model.Console, args []string) {
	buff := bytes.Buffer{}
	for _, command := range console.Commands {
		buff.WriteString(command.Name())
		buff.WriteString("\n")
		buff.WriteString("    Usage: ")
		buff.WriteString(command.Usage())
		buff.WriteString("\n")
		buff.WriteString("    ")
		buff.WriteString(command.Description())
		buff.WriteString("\n")
	}
	fmt.Println(buff.String())
}
