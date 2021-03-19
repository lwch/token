package redis

import (
	"context"
	"errors"
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
	_, err = m.cli.TxPipelined(context.Background(), func(pipe redis.Pipeliner) error {
		err = pipe.SetNX(context.Background(), tk.GetTK(), string(data), m.ttl).Err()
		if err != nil {
			return err
		}
		return pipe.SetNX(context.Background(), tk.GetUID(), tk.GetTK(), m.ttl).Err()
	})
	return err
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
	ok, err := tk.Verify([]byte(data))
	if err != nil {
		return ok, err
	}
	if ok {
		m.cli.Pipelined(context.Background(), func(pipe redis.Pipeliner) error {
			pipe.Expire(context.Background(), tk.GetTK(), m.ttl)
			pipe.Expire(context.Background(), tk.GetUID(), m.ttl)
			return nil
		})
	}
	return ok, err
}

// Revoke revoke token
func (m *Mgr) Revoke(tk string) {
	m.cli.Del(context.Background(), tk)
}

// Get get token by uid
func (m *Mgr) Get(uid string, tk token.Token) error {
	token, err := m.cli.Get(context.Background(), uid).Result()
	if err == redis.Nil {
		return errors.New("not found")
	}
	if err != nil {
		return err
	}
	data, err := m.cli.Get(context.Background(), token).Result()
	if err == redis.Nil {
		return errors.New("not found")
	}
	if err != nil {
		return err
	}
	return tk.UnSerialize(token, []byte(data))
}
