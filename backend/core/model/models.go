package models

import "time"

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Image          string    `json:"image"`
	Password       string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	NameNormalized string    `json:"-"`
}
