package models

import "time"

type State struct {
	ID        string    `json:"id"`
	Workspace string    `json:"workspace"`
	Serial    int64     `json:"serial"`
	MD5       string    `json:"md5"`
	State     []byte    `json:"state"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LockID    string    `json:"lock_id,omitempty"`
	LockedAt  time.Time `json:"locked_at,omitempty"`
}

type StateLock struct {
	ID        string    `json:"id"`
	Operation string    `json:"operation"`
	Info      string    `json:"info"`
	Who       string    `json:"who"`
	Version   string    `json:"version"`
	Created   time.Time `json:"created"`
	Path      string    `json:"path"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type TokenRequest struct {
	Code      string
	CreatedAt time.Time
}
