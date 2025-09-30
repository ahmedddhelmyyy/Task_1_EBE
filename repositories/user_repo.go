//repository hides GORM details behind an interface—DB-agnostic.
// Data-access layer. Only talks to the database (via GORM here).
//3
package repositories

import (
	"HelmyTask/models" // Import our User model to map results.

	"gorm.io/gorm" // GORM DB type is injected so repos are testable/mocked.

)

// UserRepository defines the operations our service layer expects.
// Depending on interfaces (not concrete types) helps testability and swapping implementations.
type  UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
}
 
// userRepo is a private struct implementing UserRepository.
// It holds a *gorm.DB that can connect to any dialect (mysql/postgres/sqlite/sqlserver).
type userRepo struct{ db *gorm.DB }

// NewUserRepository is a constructor that injects *gorm.DB and returns an interface.
// This allows main.go to wire dependencies without exposing concrete types to other layers.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// Create inserts a new user row using GORM's Create method.
func (r *userRepo) Create(u *models.User) error {
	return r.db.Create(u).Error // .Error exposes any DB error to caller.
}

// FindByEmail queries for a user with the given email.
// We use a parameterized query (WHERE email = ?) which GORM compiles safely for the dialect.
func (r *userRepo) FindByEmail(email string) (*models.User, error) {
	var u models.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil // Return pointer to the found user.
}

func (r *userRepo) FindByID(id uint) (*models.User, error) {
	var u models.User
	if err := r.db.First(&u, id).Error; err != nil {  // First(&u, id) loads where primary key = id.
		return nil, err
	}
	return &u, nil
}
