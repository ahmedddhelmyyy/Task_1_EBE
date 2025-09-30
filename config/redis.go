//returns a connected Redis client.use this for caching, token blacklists, rate-limit, etc.

package config

import (
	"context"  // Used to Ping Redis.
	"log"

	"github.com/redis/go-redis/v9"
)

// InitRedis creates a Redis client based on config and verifies the connection with PING.
func InitRedis(cfg *Config) *redis.Client {
	// Create a new client with address, password, and DB index from our config.
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass, // Auth password (empty if none
		DB:       cfg.RedisDB, // Logical DB index, typically 0.
	})

	// Verify connectivity by sending a PING command.
	if err := rdb.Ping(context.Background()).Err(); err != nil {// connectivity test
		log.Fatalf("[redis] ping error: %v", err)// Stop if Redis unavailable (adjust per needs).
	}
	return rdb // Return connected client for use in app.
}
