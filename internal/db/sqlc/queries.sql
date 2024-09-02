-- name: GetUser :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;

-- name: GetUserFromEmail :one
SELECT * FROM users 
WHERE email = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, pw_hash, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ChangePassword :exec
UPDATE users
SET pw_hash = $2
WHERE user_id = $1;

-- name: CreateSession :exec
INSERT INTO sessions (user_id, token, expires_at)
VALUES ($1, $2, $3);

-- name: GetSession :one
SELECT * FROM sessions
WHERE token = $1 LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token = $1;

-- name: CreateDevice :one
INSERT INTO devices (user_id, display_name)
VALUES ($1, $2)
RETURNING *;

-- name: GetDeviceByMacAddress :one
SELECT * FROM devices
WHERE mac_addr = $1 LIMIT 1;

-- name: GetDevicesByUser :many
SELECT * FROM devices
WHERE user_id = $1;

-- name: RenameDevice :exec
UPDATE devices
SET display_name = $2
WHERE device_id = $1;

-- name: UpdateDeviceMacAddress :exec
UPDATE devices
SET mac_addr = $2
WHERE device_id = $1;

-- name: CreateProvisionStaging :exec
INSERT INTO provision_staging (device_id, contract)
VALUES ($1, $2);

-- name: GetProvisionStagingByContract :one
SELECT * FROM provision_staging
WHERE contract = $1 LIMIT 1;

-- name: DeleteProvisionStaging :exec
DELETE FROM provision_staging 
WHERE device_id = $1;
