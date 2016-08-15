package api

type Command interface {
	Usage() string
	CommandName() string
	Run(arguments ...string) (bool, error)
}
