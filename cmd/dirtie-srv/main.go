package main

import (
	"fmt"
  "os"
  "os/signal"
	"syscall"

	"github.com/frozenkro/dirtie-srv/internal/api"
	"github.com/frozenkro/dirtie-srv/internal/core"
	// "github.com/frozenkro/dirtie-srv/internal/hub"

)

func main() {
	fmt.Print("Running dirtie-srv mono driver")
  
  sigChan := make(chan os.Signal, 1)
  signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

  deps := core.NewDeps()
  go api.Init(deps)
  // go hub.Init()
  
  <-sigChan

  fmt.Println("SIGTERM rcvd, shutting down")


  // TODO Any shutdown logic
}

