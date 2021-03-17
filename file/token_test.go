package file

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type tk struct {
	Token string
	Uid   string
	Name  string
}

func newToken(uid, name string) *tk {
	enc := md5.Sum([]byte(fmt.Sprintf("%d", rand.Int())))
	return &tk{
		Token: fmt.Sprintf("%x", enc),
		Uid:   uid,
		Name:  name,
	}
}

func (tk *tk) GetTK() string {
	return tk.Token
}

func (tk *tk) GetUID() string {
	return tk.Uid
}

func (tk *tk) GetName() string {
	return tk.Name
}

func (tk *tk) Serialize() ([]byte, error) {
	return json.Marshal(tk)
}

func (tk *tk) UnSerialize(token string, data []byte) error {
	return json.Unmarshal(data, tk)
}

func (token *tk) Verify(data []byte) (bool, error) {
	var dst tk
	err := json.Unmarshal([]byte(data), &dst)
	if err != nil {
		return false, err
	}
	if token.Token == dst.Token {
		*token = dst
		return true, nil
	}
	return false, nil
}

func TestFileToken(t *testing.T) {
	mgr := NewManager(os.TempDir(), time.Minute)
	tk1 := newToken("1", "hello")
	tk2 := newToken("2", "world")
	var tk3 tk

	err := mgr.Save(tk1)
	if err != nil {
		t.Fatalf("unexpected save token: %v", err)
	}
	ok, err := mgr.Verify(&tk{Token: tk1.Token})
	if err != nil {
		t.Fatalf("unexpected verify token: %v", err)
	}
	if !ok {
		t.Fatal("verify token failed: tk1")
	}
	err = mgr.Get(tk1.Uid, &tk3)
	if err != nil {
		t.Fatalf("get token by tk1.Uid failed: %v", err)
	}
	if tk1.Uid != tk3.Uid {
		t.Fatalf("unexpected uid between tk1 and tk3")
	}
	if tk1.Name != tk3.Name {
		t.Fatalf("unexpected name between tk1 and tk3")
	}
	mgr.Revoke(tk1.Token)
	ok, err = mgr.Verify(&tk{Token: tk1.Token})
	if err != nil {
		t.Fatalf("unexpected verify token: %v", err)
	}
	if ok {
		t.Fatal("unxepected verify token success: tk1")
	}

	err = mgr.Save(tk2)
	if err != nil {
		t.Fatalf("unexpected save token: %v", err)
	}
	ok, err = mgr.Verify(&tk{Token: tk2.Token})
	if err != nil {
		t.Fatalf("unexpected verify token: %v", err)
	}
	if !ok {
		t.Fatal("verify token failed: tk2")
	}
	time.Sleep(time.Minute + 10*time.Second)
	ok, err = mgr.Verify(&tk{Token: tk2.Token})
	if err != nil {
		t.Fatalf("unexpected verify token: %v", err)
	}
	if ok {
		t.Fatal("unxepected verify token success: tk2")
	}
}
