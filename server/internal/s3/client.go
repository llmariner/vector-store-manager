package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	laws "github.com/llmariner/common/pkg/aws"
	"github.com/llmariner/vector-store-manager/server/internal/config"
)

const (
	partMiBs int64 = 128
)

// NewClient returns a new S3 client.
func NewClient(ctx context.Context, c config.S3Config) (*Client, error) {
	opts := laws.NewS3ClientOptions{
		EndpointURL: c.EndpointURL,
		Region:      c.Region,
	}
	if ar := c.AssumeRole; ar != nil {
		opts.AssumeRole = &laws.AssumeRole{
			RoleARN:    ar.RoleARN,
			ExternalID: ar.ExternalID,
		}
	}
	svc, err := laws.NewS3Client(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Client{
		svc:    svc,
		bucket: c.Bucket,
	}, nil
}

// Client is a client for S3.
type Client struct {
	svc    *s3.Client
	bucket string
}

// Download downloads a S3 object and writes it to w.
func (c *Client) Download(ctx context.Context, w io.WriterAt, key string) error {
	downloader := manager.NewDownloader(c.svc, func(d *manager.Downloader) {
		d.PartSize = partMiBs * 1024 * 1024
	})
	_, err := downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}
