package services_test // Use _test to exercise the public API like an external consumer.

import ( // Imports for tests.
	"context" // For Redis calls in assertions (optional).
	"fmt" // For formatting emails in loop.
	"testing" // Go test framework.
	"time" // TTL/retention values if needed.

	"github.com/alicebob/miniredis/v2" // Fake Redis server for tests (no external dependency).
	"github.com/redis/go-redis/v9" // Redis client to talk to miniredis.

	"HelmyTask/models" // DTOs and model.
	"HelmyTask/repositories" // Repo ctor.
	"HelmyTask/services" // Service ctor.
	"HelmyTask/utils/redislog" // Redis logger used by the service.

	"gorm.io/driver/sqlite" // In-memory SQLite driver.
	"gorm.io/gorm" // GORM ORM.
)

// newTestDeps spins up in-memory DB + fake Redis + Redis logger, returns a ready service and useful handles.
func newTestDeps(t *testing.T) (services.UserService, *miniredis.Miniredis, *redis.Client) {
	t.Helper() // Mark as helper so failures point to caller line.

	// 1) Create a fresh in-memory SQLite DB for this test case.
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{}) // Lives only in memory.
	if err != nil {
		t.Fatalf("open sqlite: %v", err) // Fail test if DB cannot open.
	}
	// 2) Migrate the schema we need (User).
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate: %v", err) // Fail if migration fails.
	}
	// 3) Build the repository against this DB.
	repo := repositories.NewUserRepository(db) // Concrete repo for tests.

	// 4) Start a fake Redis server in memory (no network).
	mr := miniredis.RunT(t) // Auto-closes when test finishes.

	// 5) Create a Redis client pointed to the fake server.
	rdb := redis.NewClient(&redis.Options{ // Standard go-redis client.
		Addr: mr.Addr(), // Use address provided by miniredis (127.0.0.1:<port>).
	})

	// 6) Build a Redis logger that writes into a list key "testlogs:app".
	rlog := redislog.New(rdb, "testlogs:app", 1000, 7*24*time.Hour) // Keep last 1000; expire in 7 days.

	// 7) Construct the service with repo + Redis client + Redis logger.
	svc := services.NewUserService(repo, rdb, rlog) // This is what we will test.

	// 8) Return service + fake redis handles to allow assertions on logs.
	return svc, mr, rdb
}

func TestCreateAndGetUser_WithRedisLogging(t *testing.T) {
	// Prepare a service wired with in-memory DB + fake Redis + Redis logger.
	svc, mr, rdb := newTestDeps(t)
	_ = mr // Keep the variable to show we own the fake server lifecycle.

	// Create a new user via the service (this should log to Redis and create a DB row).
	u, err := svc.CreateUser(models.RegisterRequest{
		Name:     "ahmed", // Service should normalize this (e.g., "Ahmed").
		Email:    "ahmed@example.com", // Must be unique.
		Password: "secret123", // Will be hashed by service.
	})
	if err != nil {
		t.Fatalf("create: %v", err) // Creating should succeed.
	}
	if u.ID == 0 {
		t.Fatalf("expected non-zero ID after create")
	}

	// Read the same user back (GetUser reuses GetByID, which tries cache first).
	got, err := svc.GetUser(u.ID)
	if err != nil {
		t.Fatalf("get: %v", err) // Should be found.
	}
	if got.Email != "ahmed@example.com" {
		t.Fatalf("unexpected email: %q", got.Email) // Ensure fields persisted.
	}

	// Assert that some logs were pushed to Redis (LPUSH into "testlogs:app").
	ctx := context.Background() // Context for Redis calls.
	n, err := rdb.LLen(ctx, "testlogs:app").Result() // Count entries in Redis logs list.
	if err != nil {
		t.Fatalf("redis llen: %v", err) // Should not error against miniredis.
	}
	if n == 0 {
		t.Fatalf("expected some logs written to redis, got 0") // Ensure logging happened.
	}
}

func TestUpdateAndDeleteUser_WithRedisLogging(t *testing.T) {
	// Build service with fake redis logger as before.
	svc, _, rdb := newTestDeps(t)

	// Seed a user to update/delete.
	u, err := svc.CreateUser(models.RegisterRequest{
		Name:     "mona",
		Email:    "mona@example.com",
		Password: "pass1234",
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	// Update the user's name and password (logs should record "UpdateUser called" and cache refresh).
	newName := "Mona Lisa"
	newPass := "newpass567"
	upd, err := svc.UpdateUser(u.ID, models.UpdateUserRequest{
		Name:     &newName, // Partial update: only name + password.
		Password: &newPass, // Service will hash it.
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if upd.Name == "" {
		t.Fatalf("expected updated name, got empty")
	}

	// Delete the user (logs should record "DeleteUser called" and success).
	if err := svc.DeleteUser(u.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Ensure logs list has grown (very light assertion just to prove writes happened).
		if l, _ := rdb.LLen(context.Background(), "testlogs:app").Result(); l == 0 {
		t.Fatalf("expected redis logs after update/delete")
	}

}

func TestListUsers_WithRedisLogging(t *testing.T) {
	// Fresh service and fake redis.
	svc, _, _ := newTestDeps(t)

	// Seed multiple users.
	for i := 0; i < 5; i++ {
		_, err := svc.CreateUser(models.RegisterRequest{
			Name:     fmt.Sprintf("u%d", i), // Distinct name per user.
			Email:    fmt.Sprintf("u%d@ex.com", i), // Distinct email per user.
			Password: "p123456", // Arbitrary valid password.
		})
		if err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}

	// List page 1, limit 2 (should get exactly 2 items, total >= 5).
	page, err := svc.ListUsers(1, 2)
	if err != nil {
		t.Fatalf("list p1: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 items on page 1, got %d", len(page.Items))
	}
	if page.Total < 5 {
		t.Fatalf("expected total >= 5, got %d", page.Total)
	}
}
