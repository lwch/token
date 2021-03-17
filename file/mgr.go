package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/lwch/token"
)

// Mgr token manager
type Mgr struct {
	ttl      time.Duration
	cacheDir string
}

// DefaultTTL default ttl
const DefaultTTL = time.Hour

// NewManager new token manager
func NewManager(dir string, ttl time.Duration) *Mgr {
	ret := new(Mgr)
	ret.ttl = ttl
	ret.cacheDir = dir
	os.MkdirAll(dir, 0755)
	go func() {
		for {
			ret.clear()
			time.Sleep(time.Minute)
		}
	}()
	return ret
}

func (m *Mgr) clear() {
	files, _ := filepath.Glob(path.Join(m.cacheDir, "*.token"))
	for _, file := range files {
		fi, err := os.Stat(file)
		if err != nil {
			continue
		}
		if time.Since(fi.ModTime()) > m.ttl {
			os.Remove(file)
		}
	}
}

// Save save token
func (m *Mgr) Save(tk token.Token) error {
	data, err := tk.Serialize()
	if err != nil {
		return err
	}
	dir := path.Join(m.cacheDir, fmt.Sprintf("%s_%s.token", tk.GetUID(), tk.GetTK()))
	return ioutil.WriteFile(dir, []byte(data), 0644)
}

// Verify verify token
func (m *Mgr) Verify(tk token.Token) (bool, error) {
	files, _ := filepath.Glob(path.Join(m.cacheDir, fmt.Sprintf("*_%s.token", tk.GetTK())))
	if len(files) == 0 {
		return false, nil
	}
	data, err := ioutil.ReadFile(files[0])
	if err != nil {
		return false, err
	}
	return tk.Verify(data)
}

// Revoke revoke token
func (m *Mgr) Revoke(tk string) {
	files, _ := filepath.Glob(path.Join(m.cacheDir, fmt.Sprintf("*_%s.token", tk)))
	for _, file := range files {
		os.Remove(file)
	}
}
