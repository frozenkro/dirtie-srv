package utils

import (
	"log"
	"os"
	"time"
)

var (
  logger *log.Logger
  hasLogger bool
)

const (
  Debug = "DEBUG"
  Info = "INFO"
  Warn = "WARN"
  Error = "ERROR"
)

func getLogger() *log.Logger {
  if !hasLogger {
    logger = log.New(os.Stdout, "server: ", log.Lshortfile)
  }
  return logger
}

func Log(message string, level string) {
  logger := getLogger()
  timestamp := time.Now().UTC()

  logger.Printf("%v [%v] %v\n", timestamp, level, message)
}

func LogInfo(message string) {
  Log(message, Info)
}

func LogErr(message string) {
  Log(message, Error)
}
