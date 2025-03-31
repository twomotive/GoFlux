package main

import (
	"log"
	"os"

	"github.com/twomotive/GoFlux/internal/commands"
	"github.com/twomotive/GoFlux/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("cannot read")
	}

	programState := &commands.State{
		Cfg: cfg,
	}

	cmds := commands.Commands{
		RegisteredCommands: make(map[string]func(*commands.State, commands.Command) error),
	}
	cmds.Register("login", commands.HandlerLogin)

	if len(os.Args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
		return
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	err = cmds.Run(programState, commands.Command{Name: cmdName, Args: cmdArgs})
	if err != nil {
		log.Fatal(err)
	}
}
