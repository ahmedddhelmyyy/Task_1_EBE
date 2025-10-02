//go:build integration
// +build integration

package services_test

import (
  "os"
  "testing"

  "HelmyTask/models"
  "HelmyTask/repositories"
  "HelmyTask/services"

  "gorm.io/driver/mysql"
  "gorm.io/gorm"
)

func newMySQLService(t *testing.T) services.UserService {
  t.Helper()

  // Read DSN from env var (so you can point to any MySQL).
  dsn := os.Getenv("MYSQL_DSN")
  if dsn == "" {
    t.Skip("MYSQL_DSN not set; skipping MySQL integration test")
  }

  // Open real MySQL
  db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
  if err != nil {
    t.Fatalf("open mysql: %v", err)
  }

  // Migrate schema (creates tables as needed)
  if err := db.AutoMigrate(&models.User{}); err != nil {
    t.Fatalf("migrate: %v", err)
  }

  // Build repo & service (no Redis needed for this test, but you can add it)
  repo := repositories.NewUserRepository(db)
  svc := services.NewUserService(repo, nil, nil)

  return svc
}

func TestMySQL_CreateRead(t *testing.T) {
  svc := newMySQLService(t)

  // Clean-up note: consider truncating table before/after if you re-run often.
  u, err := svc.CreateUser(models.RegisterRequest{
    Name:     "mysql-user",
    Email:    "mysql-user@example.com",
    Password: "secret123",
  })
  if err != nil {
    t.Fatalf("create: %v", err)
  }

  got, err := svc.GetUser(u.ID)
  if err != nil {
    t.Fatalf("get: %v", err)
  }

  if got.Email != "mysql-user@example.com" {
    t.Fatalf("unexpected email: %q", got.Email)
  }
}
