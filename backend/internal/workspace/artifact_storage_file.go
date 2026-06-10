package workspace

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	appstorage "github.com/insmtx/Leros/backend/internal/infra/storage"
)

type ArtifactStorageFile struct {
	Path     string
	Filename string
	MimeType string
	FileSize int64
	Sha256   string
}

func ResolveArtifactStorageFile(ctx context.Context, orgID uint, workerID uint, storageKey string, declaredMimeType string) (*ArtifactStorageFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	st := appstorage.Get()
	bucket := appstorage.DefaultBucket()

	info, err := st.HeadObject(ctx, bucket, storageKey)
	if err != nil {
		return nil, fmt.Errorf("head artifact object: %w", err)
	}

	return &ArtifactStorageFile{
		Path:     info.Path.Path(),
		Filename: filepath.Base(info.Path.Key()),
		MimeType: detectMimeTypeFromKey(info.Path.Key(), declaredMimeType),
		FileSize: info.Size,
		Sha256:   info.ETag,
	}, nil
}

func OpenArtifactStorageFile(ctx context.Context, orgID uint, workerID uint, storageKey string) (io.ReadCloser, error) {
	st := appstorage.Get()
	bucket := appstorage.DefaultBucket()

	result, err := st.GetObject(ctx, bucket, storageKey)
	if err != nil {
		return nil, fmt.Errorf("get artifact object: %w", err)
	}
	return result.Body, nil
}

func RepoRelativePathFromStorageKey(storageKey string) string {
	key := filepath.ToSlash(strings.TrimSpace(storageKey))
	const marker = "/repo/"
	idx := strings.Index(key, marker)
	if idx < 0 {
		return ""
	}
	return strings.TrimPrefix(key[idx+len(marker):], "/")
}

func detectMimeTypeFromKey(key, declared string) string {
	if strings.TrimSpace(declared) != "" {
		return normalizeMimeType(declared)
	}
	if ext := filepath.Ext(key); ext != "" {
		if value := mime.TypeByExtension(ext); value != "" {
			return normalizeMimeType(value)
		}
	}
	return ""
}

func normalizeMimeType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if mediaType, _, err := mime.ParseMediaType(value); err == nil {
		return mediaType
	}
	if index := strings.Index(value, ";"); index >= 0 {
		return strings.TrimSpace(value[:index])
	}
	return value
}
