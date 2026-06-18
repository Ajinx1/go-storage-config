package minio

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"time"

	"github.com/minio/minio-go/v7"
)

func Upload(input UploadInput) (string, error) {
	if client == nil {
		return "", errors.New("MinIO client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, err := client.PutObject(ctx, getBucket(), input.ObjectName, input.File, input.Size, minio.PutObjectOptions{
		ContentType: input.ContentType,
	})
	if err != nil {
		return "", err
	}

	url := client.EndpointURL().String() + "/" + getBucket() + "/" + input.ObjectName

	log.Println("✅ Uploaded to:", url)

	return url, nil
}

func UploadFile(
	file multipart.File,
	size int64,
	objectName string,
	contentType string,
) (string, error) {

	input := UploadInput{
		File:        file,
		Size:        size,
		ObjectName:  objectName,
		ContentType: contentType,
	}

	return Upload(input)
}
