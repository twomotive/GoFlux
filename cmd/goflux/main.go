package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/twomotive/GoFlux/internal/commands"
	"github.com/twomotive/GoFlux/internal/config"
	"github.com/twomotive/GoFlux/internal/database"
	"github.com/twomotive/GoFlux/internal/state" // State paketini import edin
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	programState := state.New(dbQueries, cfg)

	cmds := commands.Commands{
		RegisteredCommands: make(map[string]func(*state.State, commands.Command) error),
	}
	cmds.Register("login", commands.HandlerLogin)
	cmds.Register("register", commands.HandlerRegister)
	cmds.Register("reset", commands.HandlerReset)
	cmds.Register("users", commands.HandlerGetUsers)
	cmds.Register("agg", commands.HandlerAgg)
	cmds.Register("addfeed", commands.MiddlewareLoggedIn(commands.HandlerAddFeed))
	cmds.Register("feeds", commands.HandlerGetFeeds)
	cmds.Register("follow", commands.MiddlewareLoggedIn(commands.HandlerFollow))
	cmds.Register("following", commands.MiddlewareLoggedIn(commands.HandlerFollowing))
	cmds.Register("unfollow", commands.MiddlewareLoggedIn(commands.HandlerUnfollow))

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
