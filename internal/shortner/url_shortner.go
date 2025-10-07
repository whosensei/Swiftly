package shortner

import (
	"crypto/sha256"
	"encoding/hex"
)

func Url_shorten(longurl string) string{

	hasher := sha256.New()
	hasher.Write([]byte(longurl))
	hashInbytes := hasher.Sum(nil)
	hashInstring := hex.EncodeToString(hashInbytes)
	hashInstring = hashInstring[:6]

	return hashInstring
}