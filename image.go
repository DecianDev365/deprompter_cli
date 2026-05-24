package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var mimeMap = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
}

func EncodeImage(path string) (base64Str, mimeType string, err error) {
	ext := strings.ToLower(filepath.Ext(path))
	mime, ok := mimeMap[ext]
	if !ok {
		return "", "", fmt.Errorf("unsupported image extension: %s (supported: png, jpg, jpeg, webp)", ext)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to read image: %w", err)
	}

	return base64.StdEncoding.EncodeToString(data), mime, nil
}
