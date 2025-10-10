package shortner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
)

func GenerateId() string{
	id := uuid.New()
	userId := id.String()
	return userId
}

func Url_shorten(id string,longurl string) string{
	de := fmt.Sprintf("%s%s",id,longurl)
	hasher := sha256.New()
	hasher.Write([]byte(de))
	hashInbytes := hasher.Sum(nil)
	hashInstring := hex.EncodeToString(hashInbytes)
	hashInstring = hashInstring[:6]
	return hashInstring

}