package shortener

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Base62Shortener struct{}

func NewBase62() *Base62Shortener {
	return &Base62Shortener{}
}

func (s *Base62Shortener) Generate(ctx context.Context, originalURL string) (string, error) {
	sum := sha256.Sum256([]byte(originalURL))
	num := binary.BigEndian.Uint64(sum[:8])
	return encode(num), nil
}

func encode(n uint64) string {
	if n == 0 {
		return string(alphabet[0])
	}
	var buf [11]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = alphabet[n%62]
		n /= 62
	}
	return string(buf[i:])
}
