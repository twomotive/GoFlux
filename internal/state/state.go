package state

import (
	"github.com/twomotive/GoFlux/internal/config"
	"github.com/twomotive/GoFlux/internal/database"
)

// AppState holds the global application state
type State struct {
	DB  *database.Queries
	Cfg *config.Config
}

// New creates a new AppState instance
func New(db *database.Queries, cfg *config.Config) *State {
	return &State{
		DB:  db,
		Cfg: cfg,
	}
}
