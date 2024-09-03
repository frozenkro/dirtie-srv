CREATE TABLE users (
  user_id SERIAL PRIMARY KEY,
  email VARCHAR(250) UNIQUE NOT NULL,
  name VARCHAR(250) NOT NULL,
  pw_hash BYTEA NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  last_login TIMESTAMP WITH TIME ZONE
);

CREATE TABLE sessions (
  session_id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token VARCHAR(64) NOT NULL UNIQUE,
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE devices (
  device_id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  mac_addr VARCHAR(17),
  display_name VARCHAR(250)
);

CREATE TABLE provision_staging (
  device_id INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  contract VARCHAR(64)
);