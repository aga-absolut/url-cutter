package generate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
)

// Generate генерирует случайный код из 8 символов.
func Generate(originalURL string) string {
	salt := rand.Intn(9999)
	data := sha256.Sum224([]byte(fmt.Sprintf("%s%d", originalURL, salt)))
	return hex.EncodeToString(data[:4])
}
