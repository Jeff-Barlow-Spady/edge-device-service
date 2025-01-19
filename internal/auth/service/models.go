package service

import "time"

type User struct {
    Hash      string    `json:"hash"`
    CreatedAt time.Time `json:"created_at"`
}
