package minio

import (
	"context"
	"fmt"

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

	// Print the endpoint and SSL flag for debugging
	// fmt.Println("Connecting to MinIO at:", endpoint)
	//fmt.Println("Bucket is:", getBucket())
	// fmt.Printf("Key stuffs: access %v and secret %v", config.AccessKeyID, config.SecretAccessKey)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(context.Background(), getBucket())
	if err != nil {
		return nil, err
	}

	if !exists {
		err := client.MakeBucket(context.Background(), getBucket(), minio.MakeBucketOptions{})
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

		err = client.SetBucketPolicy(context.Background(), getBucket(), publicPolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to set public policy: %w", err)
		}
	}

	return client, nil
}
