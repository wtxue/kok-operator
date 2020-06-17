package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io/ioutil"
	"os"
)

func Sha256WithFile(filename string) (string, error) {
	h := sha256.New()
	return SumWithFile(h, filename)
}

func SumWithFile(h hash.Hash, filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return Sum(h, data), nil
}

func Sum(h hash.Hash, data []byte) string {
	return hex.EncodeToString(h.Sum(data))
}
