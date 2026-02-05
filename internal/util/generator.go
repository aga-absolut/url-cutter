package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
)

func Generate(originalURL string) string {
	salt := rand.Intn(9999)
	data := sha256.Sum224([]byte(fmt.Sprintf("%s%d", originalURL, salt)))
	return hex.EncodeToString(data[:4])
}
