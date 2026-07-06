package minio

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var client *minio.Client
var bucket string

func Init(appBucket string, theConfig MinioConfig) error {
	config := LoadMinioConfigFromEnv(theConfig)
	bucket = appBucket

	c, err := Connect(config)
	if err != nil {
		return err
	}

	client = c
	return nil
}

func getBucket() string {
	return bucket
}

func Connect(config MinioConfig) (*minio.Client, error) {
	endpoint := fmt.Sprintf("%s:%d", config.Url, config.Port)

	c, err := minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure:    config.UseSSL,
		Transport: minioTransport(),
	})
	if err != nil {
		return nil, err
	}

	exists, err := c.BucketExists(context.Background(), getBucket())
	if err != nil {
		return nil, err
	}

	if !exists {
		err := c.MakeBucket(context.Background(), getBucket(), minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}

		publicPolicy := fmt.Sprintf(`{
			"Version":"2012-10-17",
			"Statement":[{
				"Effect":"Allow",
				"Principal":"*",
				"Action":["s3:GetObject"],
				"Resource":["arn:aws:s3:::%s/*"]
			}]
		}`, getBucket())

		err = c.SetBucketPolicy(context.Background(), getBucket(), publicPolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to set public policy: %w", err)
		}
	}

	return c, nil
}

func minioTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
	}
}
