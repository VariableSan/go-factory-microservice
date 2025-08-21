package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Client wraps redis client with additional functionality
type Client struct {
	*redis.Client
}

// NewClient creates a new Redis client
func NewClient(addr, password string, db int) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &Client{Client: rdb}
}

// NewClientFromURL creates a new Redis client from URL
func NewClientFromURL(url string) (*Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opt)
	return &Client{Client: rdb}, nil
}

// Ping tests the connection
func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// SetWithExpiry sets a key-value pair with expiration
func (c *Client) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	return c.Client.Set(ctx, key, value, expiry).Err()
}

// GetString gets a string value by key
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Delete deletes a key
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.Client.Del(ctx, keys...).Err()
}

// Exists checks if key exists
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.Client.Exists(ctx, keys...).Result()
}

// SetNX sets key-value only if key doesn't exist
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiry time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, expiry).Result()
}

// Increment increments a key's value
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

// IncrementWithExpiry increments and sets expiry
func (c *Client) IncrementWithExpiry(ctx context.Context, key string, expiry time.Duration) (int64, error) {
	pipe := c.Client.TxPipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiry)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	
	return incrCmd.Val(), nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.Client.Close()
}
