package minio

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
)

func Upload(input UploadInput) (string, error) {
	if client == nil {
		return "", errors.New("MinIO client not initialized")
	}

	_, err := client.PutObject(
		context.Background(),
		getBucket(),
		input.ObjectName,
		input.File,
		input.Size,
		minio.PutObjectOptions{ContentType: input.ContentType},
	)
	if err != nil {
		return "", err
	}

	url := client.EndpointURL().String() + "/" + getBucket() + "/" + input.ObjectName
	log.Println("✅ Uploaded to:", url)
	return url, nil
}

func Delete(input DeleteInput) error {
	if client == nil {
		return errors.New("MinIO client not initialized")
	}

	return client.RemoveObject(context.Background(), getBucket(), input.ObjectName,
		minio.RemoveObjectOptions{ForceDelete: true})
}

func List(input ListInput) ([]string, error) {
	if client == nil {
		return nil, errors.New("MinIO client not initialized")
	}

	var objects []string
	ctx := context.Background()

	objectCh := client.ListObjects(ctx, getBucket(), minio.ListObjectsOptions{
		Prefix:    input.Prefix,
		Recursive: true,
	})

	for obj := range objectCh {
		if obj.Err != nil {
			return nil, obj.Err
		}
		objects = append(objects, obj.Key)
	}

	return objects, nil
}

func UploadFile(file multipart.File, objectName, contentType string) (string, error) {
	var buffer []byte
	var size int64

	for {
		buf := make([]byte, 1024)
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		buffer = append(buffer, buf[:n]...)
		size += int64(n)
	}

	input := UploadInput{
		File:        bytes.NewReader(buffer),
		ObjectName:  objectName,
		Size:        size,
		ContentType: contentType,
	}

	return Upload(input)
}
