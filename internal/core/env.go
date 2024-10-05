package core

import ("os")

type Env struct {
  influx_org string
  influx_default_bucket string
  influx_token string
  influx_username string
  influx_password string

  postgres_server string
  postgres_db string
  postgres_user string
  postgres_password string

  mosquitto_uri string
  app_host string
  assets_dir string
}

func NewTestEnv() *Env {
  env := NewEnv()
  env.postgres_server = os.Getenv("POSTGRES_TEST_SERVER")
  return env
}

func NewEnv() *Env {
  var env Env

  env.influx_org = os.Getenv("INFLUX_ORG")
  env.influx_default_bucket = os.Getenv("INFLUX_DEFAULT_BUCKET")
  env.influx_token = os.Getenv("INFLUX_TOKEN")
  env.influx_username = os.Getenv("INFLUX_USERNAME")
  env.influx_password = os.Getenv("INFLUX_PASSWORD")
  env.postgres_server = os.Getenv("POSTGRES_SERVER")
  env.postgres_db = os.Getenv("POSTGRES_DB")
  env.postgres_user = os.Getenv("POSTGRES_USER")
  env.postgres_password = os.Getenv("POSTGRES_PASSWORD")
  env.mosquitto_uri = os.Getenv("MOSQUITTO_URI")
  env.app_host = os.Getenv("ASSETS_DIR")
  return &env
}
