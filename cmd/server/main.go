package main

import (
	"airline-tracker/internal/app"
)

func main() {
	app := app.NewApp()
	app.Run()
}
