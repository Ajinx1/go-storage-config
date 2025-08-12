package redis

import (
	"time"
)

type Client struct {
	conn *RedisConn
}

func NewClient(conn *RedisConn) *Client {
	return &Client{conn: conn}
}

func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return c.conn.Client.Set(c.conn.Ctx, key, value, expiration).Err()
}

func (c *Client) Get(key string) (string, error) {
	return c.conn.Client.Get(c.conn.Ctx, key).Result()
}

func (c *Client) Del(keys ...string) error {
	return c.conn.Client.Del(c.conn.Ctx, keys...).Err()
}

func (c *Client) Exists(keys ...string) (int64, error) {
	return c.conn.Client.Exists(c.conn.Ctx, keys...).Result()
}

func (c *Client) Expire(key string, expiration time.Duration) error {
	return c.conn.Client.Expire(c.conn.Ctx, key, expiration).Err()
}

func (c *Client) TTL(key string) (time.Duration, error) {
	return c.conn.Client.TTL(c.conn.Ctx, key).Result()
}

func (c *Client) LPush(key string, value interface{}) error {
	return c.conn.Client.LPush(c.conn.Ctx, key, value).Err()
}

func (c *Client) BRPop(timeout time.Duration, key string) ([]string, error) {
	return c.conn.Client.BRPop(c.conn.Ctx, timeout, key).Result()
}
