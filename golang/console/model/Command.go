package model

type Command interface {
	Name() string
	Usage() string
	Description() string
	Run(*Console, []string)
}
