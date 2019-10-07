package commands

import "github.com/saichler/habitat/golang/console/model"

type Connect struct{}

func (cmd *Connect) Name() string {
	return "connect"
}

func (cmd *Connect) Usage() string {
	return "connect <ip>"
}

func (cmd *Connect) Description() string {
	return "Connect to adjacent machine."
}

func (cmd *Connect) Run(console *model.Console, args []string) {

}
