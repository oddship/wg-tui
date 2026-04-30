package main

import (
	"log"

	"github.com/oddship/wg-tui/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
