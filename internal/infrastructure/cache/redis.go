package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"Finance-Manager-System/configs"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	redisClient *redis.Client
	enabled     bool
	ttl         time.Duration
}

type ResponsePayload struct {
	StatusCode  int    `json:"status_code"`
	ContentType string `json:"content_type"`
	Body        string `json:"body"`
}

func NewRedisClient(cfg configs.RedisConfig) (*Client, error) {
	addr := cfg.Host + ":" + cfg.Port
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return &Client{enabled: false, ttl: time.Duration(cfg.TTL) * time.Second}, err
	}

	ttl := time.Duration(cfg.TTL) * time.Second
	if ttl <= 0 {
		ttl = 120 * time.Second
	}

	return &Client{
		redisClient: rdb,
		enabled:     true,
		ttl:         ttl,
	}, nil
}

func (c *Client) Enabled() bool {
	return c != nil && c.enabled && c.redisClient != nil
}

func (c *Client) TTL() time.Duration {
	if c == nil {
		return 120 * time.Second
	}
	return c.ttl
}

func (c *Client) BuildRequestKey(userID uuid.UUID, method string, path string, rawQuery string) string {
	if rawQuery == "" {
		return "cache:user:" + userID.String() + ":" + method + ":" + path
	}
	return "cache:user:" + userID.String() + ":" + method + ":" + path + "?" + rawQuery
}

func (c *Client) BuildUserPrefix(userID uuid.UUID) string {
	return "cache:user:" + userID.String() + ":"
}

func (c *Client) GetResponse(ctx context.Context, key string) (*ResponsePayload, bool, error) {
	if !c.Enabled() {
		return nil, false, nil
	}
	raw, err := c.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var payload ResponsePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, false, err
	}
	return &payload, true, nil
}

func (c *Client) SetResponse(ctx context.Context, key string, payload ResponsePayload) error {
	if !c.Enabled() {
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.redisClient.Set(ctx, key, b, c.ttl).Err()
}

func (c *Client) InvalidateByUser(ctx context.Context, userID uuid.UUID) error {
	if !c.Enabled() {
		return nil
	}
	pattern := c.BuildUserPrefix(userID) + "*"
	var cursor uint64
	for {
		keys, nextCursor, err := c.redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.redisClient.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

func (c *Client) Stats(ctx context.Context) (map[string]string, error) {
	if !c.Enabled() {
		return map[string]string{"enabled": "false"}, nil
	}
	info := map[string]string{
		"enabled": "true",
		"db":      strconv.Itoa(c.redisClient.Options().DB),
		"addr":    c.redisClient.Options().Addr,
		"ttl":     fmt.Sprintf("%ds", int(c.ttl.Seconds())),
	}
	return info, nil
}
