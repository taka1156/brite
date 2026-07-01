package storage

import (
	"context"
	"io"
)

type Storage interface {
	Upload(ctx context.Context, bucket, key string, body io.Reader, contentType string) error
	Delete(ctx context.Context, bucket, key string) error
}
