package helper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrFileTooLarge  = errors.New("file size exceeds the limit")
	ErrInvalidFormat = errors.New("file format is not allowed")
)

func UploadFile(r *http.Request, field string, allowedExts []string, maxSizeMB int64) (*string, error) {
	file, header, err := r.FormFile(field)
	{
		if err != nil || file == nil {
			return nil, nil
		}
	}

	defer file.Close()

	if header.Size > maxSizeMB*1024*1024 {
		return nil, fmt.Errorf("%w: max %dMB", ErrFileTooLarge, maxSizeMB)
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))

	if !isAllowedExt(ext, allowedExts) {
		return nil, fmt.Errorf("%w: allowed %v", ErrInvalidFormat, allowedExts)
	}

	now := time.Now()
	dir := filepath.Join("uploads", now.Format("2006"), now.Format("01"), now.Format("02"))

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%d_%s", now.UnixNano(), sanitizeFilename(header.Filename))
	destPath := filepath.Join(dir, filename)

	dst, err := os.Create(destPath)
	{
		if err != nil {
			return nil, err
		}
	}

	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	path := filepath.ToSlash(destPath)
	return &path, nil
}

func isAllowedExt(ext string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(ext, a) {
			return true
		}
	}
	return false
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	return name
}
