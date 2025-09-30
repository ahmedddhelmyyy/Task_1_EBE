// GORM model + simple DTOs used in handlers.

package models

import "time"

//user represents a user record in the database 
//Gorm tags configure primary key , sizes and constrains
//json tags control how fields serialized in api respone 
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:120;not null" json:"name"`
	Email     string    `gorm:"size:180;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"` // hashed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DTOs (request/response)
// RegisterRequest is the expected payload for the register endpoint.
// Gin's binding tags add basic validation rules automatically.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

//expectedd payload for the login endpoint
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

//small resonse object hodl jwt token 
type AuthResponse struct {
	Token string `json:"token"`
}
