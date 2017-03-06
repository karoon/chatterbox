package auth

import (
	"crypto/sha256"
	"fmt"
)

// Password type
type Password string

// Hash returns hash string of raw password
func (p *Password) Hash() string {
	passHash := sha256.New()
	passHash.Write([]byte(*p))
	h := fmt.Sprintf("%x", passHash.Sum(nil))
	return string(h)
}

func (p *Password) String() string {
	return string(*p)
}
