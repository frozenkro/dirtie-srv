package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/frozenkro/dirtie-srv/internal/api"
	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/di"
	"github.com/frozenkro/dirtie-srv/internal/hub"
)

func main() {
	fmt.Print("Running dirtie-srv mono driver\n")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	core.SetupEnv()
	deps := di.NewDeps(context.Background())

	go api.Init(deps)
	go hub.Init(deps)

	<-sigChan

	fmt.Println("SIGTERM rcvd, shutting down")
}
