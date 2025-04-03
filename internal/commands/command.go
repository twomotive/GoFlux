package commands

import (
	"errors"

	"github.com/twomotive/GoFlux/internal/state"
)

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	RegisteredCommands map[string]func(*state.State, Command) error
}

func (c *Commands) Register(name string, f func(*state.State, Command) error) {
	c.RegisteredCommands[name] = f
}

func (c *Commands) Run(s *state.State, cmd Command) error {
	f, ok := c.RegisteredCommands[cmd.Name]
	if !ok {
		return errors.New("command not found")
	}
	return f(s, cmd)
}
