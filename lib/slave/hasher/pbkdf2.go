package hasher

import (
	"crypto/sha256"

	"github.com/aaronang/cong-the-ripper/lib"
	"golang.org/x/crypto/pbkdf2"
)

type Pbkdf2 struct {
}

func (p *Pbkdf2) Hash(candidate []byte, task *lib.Task) []byte {
	return pbkdf2.Key(candidate, task.Salt, task.Iter, sha256.Size, sha256.New)
}
