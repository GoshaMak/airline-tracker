package main

import (
	"airline-tracker/internal/server"
	"os"
)

func main() {
	srv := server.NewServer()
	code := srv.Run()
	os.Exit(code)
}
