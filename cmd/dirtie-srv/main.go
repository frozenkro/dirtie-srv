package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/frozenkro/dirtie-srv/internal/api"
	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/di"
	"github.com/frozenkro/dirtie-srv/internal/hub"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Print("Running dirtie-srv mono driver\n")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if os.Getenv("APP_HOST") != "container" {
		err := godotenv.Load("../../.env")
		if err != nil {
      err = godotenv.Load("./.env")
      if err != nil {
        panic("Unable to locate .env\n")
      }
		}
	}

  core.SetupEnv()
	deps := di.NewDeps()

	go api.Init(deps)
	go hub.Init()

	<-sigChan

	fmt.Println("SIGTERM rcvd, shutting down\n")
}
