package core

import (
  "os"
  "regexp"

  "github.com/joho/godotenv"
)

var (
  INFLUX_ORG string
  INFLUX_DEFAULT_BUCKET string
  INFLUX_TOKEN string
  INFLUX_USERNAME string
  INFLUX_PASSWORD string
  INFLUX_URI string

  POSTGRES_SERVER string
  POSTGRES_DB string
  POSTGRES_USER string
  POSTGRES_PASSWORD string

  MOSQUITTO_URI string
  APP_HOST string
  ASSETS_DIR string
  DIRTIE_ENV string
  DOMAIN string
  SENDGRID_API_KEY string
)

func ProjectRootDir() string {
    re := regexp.MustCompile(`^(.*` + PROJECT_DIR_NAME + `)`)
    cwd, _ := os.Getwd()
    return string(re.Find([]byte(cwd)))
}

func SetupTestEnv() {
  SetupEnv()
  POSTGRES_SERVER = os.Getenv("POSTGRES_TEST_SERVER")
}

func SetupEnv() {
	if os.Getenv("APP_HOST") != "container" {
		err := godotenv.Load(ProjectRootDir() + `/.env`)
    if err != nil {
      panic("Unable to locate .env\n")
		}
	}

  INFLUX_ORG = os.Getenv("INFLUX_ORG")
  INFLUX_DEFAULT_BUCKET = os.Getenv("INFLUX_DEFAULT_BUCKET")
  INFLUX_TOKEN = os.Getenv("INFLUX_TOKEN")
  INFLUX_USERNAME = os.Getenv("INFLUX_USERNAME")
  INFLUX_PASSWORD = os.Getenv("INFLUX_PASSWORD")
  INFLUX_URI = os.Getenv("INFLUX_URI")
  if INFLUX_URI == "" {
    INFLUX_URI = "localhost:8086"
  }

  POSTGRES_SERVER = os.Getenv("POSTGRES_SERVER")
  POSTGRES_DB = os.Getenv("POSTGRES_DB")
  POSTGRES_USER = os.Getenv("POSTGRES_USER")
  POSTGRES_PASSWORD = os.Getenv("POSTGRES_PASSWORD")

  MOSQUITTO_URI = os.Getenv("MOSQUITTO_URI")
  APP_HOST = os.Getenv("APP_HOST")
  ASSETS_DIR = os.Getenv("ASSETS_DIR")
  DIRTIE_ENV = os.Getenv("DIRTIE_ENV")
  DOMAIN = os.Getenv("DOMAIN")
  SENDGRID_API_KEY = os.Getenv("SENDGRID_API_KEY")
}
