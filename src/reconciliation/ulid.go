package reconciliation

import (
	"crypto/rand"
	"math/big"
	"time"
)

const ulidChars = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// GenerateULID produces a standard 26-character Base32 ULID matching Python's generate_ulid().
func GenerateULID() string {
	millis := time.Now().UnixMilli()

	timePart := make([]byte, 10)
	for i := 9; i >= 0; i-- {
		timePart[i] = ulidChars[millis%32]
		millis /= 32
	}

	entropyPart := make([]byte, 16)
	maxVal := big.NewInt(int64(len(ulidChars)))
	for i := 0; i < 16; i++ {
		idx, err := rand.Int(rand.Reader, maxVal)
		if err != nil {
			entropyPart[i] = ulidChars[time.Now().UnixNano()%32]
		} else {
			entropyPart[i] = ulidChars[idx.Int64()]
		}
	}

	return string(timePart) + string(entropyPart)
}
