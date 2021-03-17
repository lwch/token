package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lwch/token"
)

// RedisConf redis config
type RedisConf struct {
	Host     string
	Port     uint16
	Password string
	DB       int
}

// Mgr token manager
type Mgr struct {
	cli *redis.Client
	ttl time.Duration
}

// DefaultTTL default ttl
const DefaultTTL = time.Hour

// NewManager new token manager
func NewManager(cfg RedisConf, ttl time.Duration) *Mgr {
	ret := new(Mgr)
	ret.cli = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ret.ttl = ttl
	return ret
}

// Save save token
func (m *Mgr) Save(tk token.Token) error {
	data, err := tk.Serialize()
	if err != nil {
		return err
	}
	_, err = m.cli.SetEX(context.Background(), tk.GetTK(), string(data), m.ttl).Result()
	if err != nil {
		return err
	}
	return nil
}

// Verify verify token
func (m *Mgr) Verify(tk token.Token) (bool, error) {
	data, err := m.cli.Get(context.Background(), tk.GetTK()).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return tk.Verify([]byte(data))
}

// Revoke revoke token
func (m *Mgr) Revoke(tk string) {
	m.cli.Del(context.Background(), tk)
}
