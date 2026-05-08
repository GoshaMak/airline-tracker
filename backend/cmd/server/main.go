package main

import (
	"airline-tracker/internal/app"
	"os"
)

func main() {
	app := app.NewApp()
	code := app.Run()
	os.Exit(code)
}
