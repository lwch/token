package token

// Token token struct
type Token interface {
	GetTK() string
	GetUID() string
	GetName() string

	Serialize() ([]byte, error)
	Verify([]byte) (bool, error)
}
