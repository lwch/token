package token

// Token token struct
type Token interface {
	GetTK() string
	GetUID() string
	GetName() string

	Load(from interface{}) error
	Save(to interface{}) error
}
