package main

import (
	"fmt"

	"github.com/twomotive/GoFlux/internal/config"
)

func main() {
	newConfig, err := config.Read()
	if err != nil {
		fmt.Print("cannot read")
	}

	newConfig.SetUser("talha")

	bar, err := config.Read()
	if err != nil {
		fmt.Print("cannot read")
	}

	fmt.Printf("%+v\n", *bar)

}
