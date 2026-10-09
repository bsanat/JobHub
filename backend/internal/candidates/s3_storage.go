package candidates

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type S3FileStore struct {
	client *s3.Client
	bucket string
}

func NewS3FileStore(endpoint, region, accessKey, secretKey, bucket string, client *http.Client) (*S3FileStore, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "https" && !(u.Scheme == "http" && u.Hostname() == "127.0.0.1")) || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, fmt.Errorf("AWS_ENDPOINT_URL_S3 must be an HTTPS origin")
	}
	if region == "" || accessKey == "" || secretKey == "" || !regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`).MatchString(bucket) {
		return nil, fmt.Errorf("S3 region, credentials and valid bucket are required")
	}
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	sdk := s3.NewFromConfig(aws.Config{Region: region, Credentials: credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""), HTTPClient: client, RetryMaxAttempts: 2, RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired, ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(strings.TrimRight(endpoint, "/"))
		o.UsePathStyle = true
	})
	return &S3FileStore{client: sdk, bucket: bucket}, nil
}

func (s *S3FileStore) Put(ctx context.Context, key string, source io.Reader) error {
	if !storageKeyPattern.MatchString(key) {
		return fmt.Errorf("invalid storage key")
	}
	// The upload service caps PDFs at 10 MiB; a seekable body is required for SigV4.
	body, err := io.ReadAll(io.LimitReader(source, MaxResumeBytes+1))
	if err != nil {
		return err
	}
	if int64(len(body)) > MaxResumeBytes {
		return fmt.Errorf("resume exceeds storage limit")
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), Body: bytes.NewReader(body), ContentType: aws.String("application/pdf"), IfNoneMatch: aws.String("*")})
	return s3Error("upload", err)
}

func (s *S3FileStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if !storageKeyPattern.MatchString(key) {
		return nil, fmt.Errorf("invalid storage key")
	}
	res, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, s3Error("download", err)
	}
	return res.Body, nil
}

func (s *S3FileStore) Delete(ctx context.Context, key string) error {
	if !storageKeyPattern.MatchString(key) {
		return fmt.Errorf("invalid storage key")
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	return s3Error("delete", err)
}

func s3Error(operation string, err error) error {
	if err == nil {
		return nil
	}
	var api smithy.APIError
	if errors.As(err, &api) && api.ErrorCode() == "NoSuchKey" {
		return os.ErrNotExist
	}
	// SDK error messages can contain private object paths and credentials.
	return fmt.Errorf("%s resume: %w", operation, ErrFileStorageUnavailable)
}
