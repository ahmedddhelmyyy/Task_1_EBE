// implements use-cases (register, login, get) and issues JWTs.
// No HTTP or GORM details leak out.
// 2
package services

import (
	"errors"
	"time"

	"HelmyTask/core"         // Use domain/pure logic (NormalizeName).
	"HelmyTask/models"       //// DTOs and User model types.
	"HelmyTask/repositories" // Interface for persistence operations.
	"HelmyTask/utils"        //// Password hashing utilities.

	"github.com/golang-jwt/jwt/v5" // Library to create JWT tokens with claims.
)

// UserService defines use-cases the handler layer will call.
// It returns domain types (User) or tokens, not HTTP responses.
type UserService interface {
	Register(req models.RegisterRequest) (*models.User, error)
	Login(req models.LoginRequest, jwtSecret string, exp time.Duration) (string, error) // Validates creds & returns Jwt
	GetByID(id uint) (*models.User, error)                                              // Fetch a user by ID for "me" endpoint.
}

// userService is a concrete implementation that depends on a repository.
type userService struct {
	// Data access is abstracted behind this interface.
	repo repositories.UserRepository
}

// NewUserService constructs a service with a repository dependency.
func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

// Register validates uniqueness, hashes the password, applies domain rules, and saves the user.
func (s *userService) Register(req models.RegisterRequest) (*models.User, error) {
	// Check if email already exists; if repository returns nil error, a user exists.
	if _, err := s.repo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already exists")
	}
	// Hash the plaintext password for secure storage.
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	// Build a User entity applying domain normalization on the name.
	u := &models.User{
		Name:     core.NormalizeName(req.Name),
		Email:    req.Email,
		Password: hash,
	}
	// Persist the user via repository.
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return u, nil // Return the created user (without password in JSON because of json:"-").
}

// Login verifies the user’s credentials and returns a signed JWT if valid.
func (s *userService) Login(req models.LoginRequest, jwtSecret string, exp time.Duration) (string, error) {
	// Look up user by email; if not found or other error, hide details for security.
	u, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	// Verify plaintext password against stored bcrypt hash.
	if !utils.CheckPassword(u.Password, req.Password) {
		return "", errors.New("invalid credentials")
	}

	// Create standard JWT claims (subject, issued-at, expiration).
	claims := jwt.MapClaims{
		"sub": u.ID,                       // Subject: the authenticated user’s ID.
		"exp": time.Now().Add(exp).Unix(), // Expiration timestamp (seconds).
		"iat": time.Now().Unix(),          // Issued-at timestamp.
		"eml": u.Email,                    // Optional custom claim (email).

	}
	// Sign with HS256 using our secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret)) // Return compact string form of token (or error).
}

// GetByID fetches a user entity so the handler can return current user info.
func (s *userService) GetByID(id uint) (*models.User, error) {
	return s.repo.FindByID(id)
}
