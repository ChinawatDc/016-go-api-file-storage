package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ioCopy(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}

type saJSON struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}

func readServiceAccount(path string) (accessID string, privateKey []byte, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	var sa saJSON
	if err := json.Unmarshal(b, &sa); err != nil {
		return "", nil, err
	}
	if sa.ClientEmail == "" || sa.PrivateKey == "" {
		return "", nil, fmt.Errorf("invalid service account json")
	}
	return sa.ClientEmail, []byte(sa.PrivateKey), nil
}
