package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func ExtLower(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	return strings.TrimPrefix(ext, ".")
}

func IsAllowedExt(ext string, allowed map[string]bool) bool {
	if ext == "" {
		return false
	}
	return allowed[strings.ToLower(ext)]
}

func RandomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func BuildSafeFilename(originalName string) (string, error) {
	base := filepath.Base(originalName)
	base = sanitize(base)

	r, err := RandomHex(6)
	if err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s_%s", ts, r, base), nil
}

func sanitize(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "..", ".")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	return name
}
