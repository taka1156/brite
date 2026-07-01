package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Storage struct {
	client *s3.Client
}

func NewR2Storage() (Storage, error) {
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	endpoint := os.Getenv("R2_ENDPOINT")
	usePathStyle, err := strconv.ParseBool(os.Getenv("ENABLE_PATH_STYLE_ENDPOINTS"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ENABLE_PATH_STYLE_ENDPOINTS: %w", err)
	}

	if accessKey == "" || secretKey == "" || endpoint == "" {
		return nil, fmt.Errorf("R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, and R2_ENDPOINT must be set")
	}

	cfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = usePathStyle
	})

	return &R2Storage{client: client}, nil
}

func (r *R2Storage) Upload(ctx context.Context, bucket, key string, body io.Reader, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		Body:         body,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to R2: %w", err)
	}
	fmt.Printf("Uploaded %s to R2 bucket %s\n", key, bucket)
	return nil
}

func (r *R2Storage) Delete(ctx context.Context, bucket, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from R2: %w", err)
	}
	fmt.Printf("Deleted %s from R2 bucket %s\n", key, bucket)
	return nil
}
