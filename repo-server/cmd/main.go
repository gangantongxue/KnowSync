package main

import (
	"github.com/gangantongxue/knowsync/repo-server/internal/app"
)

func main() {
	if err := app.NewApp(); err != nil {
		panic(err)
	}
}
