package redis

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lwch/token"
)

// RedisConf redis config
type RedisConf struct {
	Addrs    []string
	User     string
	Password string
	DB       int
}

// Mgr token manager
type Mgr struct {
	cli        *redis.Client
	clusterCli *redis.ClusterClient
	ttl        time.Duration
}

// DefaultTTL default ttl
const DefaultTTL = time.Hour

// NewManager new token manager
func NewManager(cfg RedisConf, ttl time.Duration) *Mgr {
	ret := new(Mgr)
	if len(cfg.Addrs) > 1 {
		ret.clusterCli = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: cfg.Addrs,
			NewClient: func(opt *redis.Options) *redis.Client {
				opt.DB = cfg.DB
				return redis.NewClient(opt)
			},
			Username: cfg.User,
			Password: cfg.Password,
		})
	} else {
		ret.cli = redis.NewClient(&redis.Options{
			Addr:     cfg.Addrs[0],
			Username: cfg.User,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	}
	ret.ttl = ttl
	return ret
}

// Save save token
func (m *Mgr) Save(tk token.Token) error {
	data, err := tk.Serialize()
	if err != nil {
		return err
	}
	pipe := func(pipe redis.Pipeliner) error {
		err = pipe.SetNX(context.Background(), tk.GetTK(), string(data), m.ttl).Err()
		if err != nil {
			return err
		}
		return pipe.SetNX(context.Background(), tk.GetUID(), tk.GetTK(), m.ttl).Err()
	}
	if m.cli != nil {
		_, err = m.cli.TxPipelined(context.Background(), pipe)
	} else {
		_, err = m.clusterCli.TxPipelined(context.Background(), pipe)
	}
	return err
}

// Verify verify token
func (m *Mgr) Verify(tk token.Token) (bool, error) {
	var data string
	var err error
	if m.cli != nil {
		data, err = m.cli.Get(context.Background(), tk.GetTK()).Result()
	} else {
		data, err = m.clusterCli.Get(context.Background(), tk.GetTK()).Result()
	}
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
		pipe := func(pipe redis.Pipeliner) error {
			pipe.Expire(context.Background(), tk.GetTK(), m.ttl)
			pipe.Expire(context.Background(), tk.GetUID(), m.ttl)
			return nil
		}
		if m.cli != nil {
			m.cli.Pipelined(context.Background(), pipe)
		} else {
			m.clusterCli.Pipelined(context.Background(), pipe)
		}
	}
	return ok, err
}

// Revoke revoke token
func (m *Mgr) Revoke(tk string) {
	m.cli.Del(context.Background(), tk)
}

// Get get token by uid
func (m *Mgr) Get(uid string, tk token.Token) error {
	var token string
	var err error
	if m.cli != nil {
		token, err = m.cli.Get(context.Background(), uid).Result()
	} else {
		token, err = m.clusterCli.Get(context.Background(), uid).Result()
	}
	if err == redis.Nil {
		return errors.New("not found")
	}
	if err != nil {
		return err
	}
	var data string
	if m.cli != nil {
		data, err = m.cli.Get(context.Background(), token).Result()
	} else {
		data, err = m.clusterCli.Get(context.Background(), token).Result()
	}
	if err == redis.Nil {
		return errors.New("not found")
	}
	if err != nil {
		return err
	}
	return tk.UnSerialize(token, []byte(data))
}
