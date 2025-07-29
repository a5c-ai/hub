package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	config2 "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithy "github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// S3Backend implements the Backend interface using S3-compatible storage
type S3Backend struct {
	config    S3Config
	client    *s3.Client
	presigner *s3.PresignClient
}

// NewS3Backend creates a new S3-compatible storage backend
func NewS3Backend(config S3Config) (*S3Backend, error) {
	if config.Bucket == "" {
		return nil, fmt.Errorf("s3 bucket name is required")
	}
	if config.AccessKeyID == "" {
		return nil, fmt.Errorf("s3 access key ID is required")
	}
	if config.SecretAccessKey == "" {
		return nil, fmt.Errorf("s3 secret access key is required")
	}

	// Initialize AWS SDK v2 configuration
	awsCfg, err := config2.LoadDefaultConfig(context.Background(),
		config2.WithRegion(config.Region),
		config2.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Configure S3 client with optional custom endpoint and path-style addressing
	clientOpts := []func(*s3.Options){
		func(o *s3.Options) {
			if config.EndpointURL != "" {
				o.EndpointResolver = s3.EndpointResolverFromURL(config.EndpointURL)
				o.UsePathStyle = true
			}
		},
	}
	client := s3.NewFromConfig(awsCfg, clientOpts...)
	presigner := s3.NewPresignClient(client)
	return &S3Backend{
		config:    config,
		client:    client,
		presigner: presigner,
	}, nil
}

// Upload uploads a file to S3-compatible storage
func (s *S3Backend) Upload(ctx context.Context, path string, reader io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
		Body:   reader,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object %s: %w", path, err)
	}
	return nil
}

// Download downloads a file from S3-compatible storage
func (s *S3Backend) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object %s: %w", path, err)
	}
	return resp.Body, nil
}

// Delete deletes a file from S3-compatible storage
func (s *S3Backend) Delete(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", path, err)
	}
	return nil
}

// Exists checks if a file exists in S3-compatible storage
func (s *S3Backend) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
			return false, nil
		}
		var respErr *smithyhttp.ResponseError
		if errors.As(err, &respErr) && respErr.Response.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check existence of object %s: %w", path, err)
	}
	return true, nil
}

// GetSize returns the size of a file in bytes
func (s *S3Backend) GetSize(ctx context.Context, path string) (int64, error) {
	resp, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get size of object %s: %w", path, err)
	}
	return *resp.ContentLength, nil
}

// GetLastModified returns the last modified time of a file
func (s *S3Backend) GetLastModified(ctx context.Context, path string) (time.Time, error) {
	resp, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last modified of object %s: %w", path, err)
	}
	return *resp.LastModified, nil
}

// List lists files with the given prefix
func (s *S3Backend) List(ctx context.Context, prefix string) ([]string, error) {
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.config.Bucket),
		Prefix: aws.String(prefix),
	})
	var keys []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects with prefix %s: %w", prefix, err)
		}
		for _, obj := range page.Contents {
			if obj.Key != nil {
				keys = append(keys, *obj.Key)
			}
		}
	}
	return keys, nil
}

// GetURL returns a presigned URL for downloading
func (s *S3Backend) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	presignResult, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to presign GetObject request for %s: %w", path, err)
	}
	return presignResult.URL, nil
}
