package minio

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

var templateCache = gocache.New(
	72*time.Hour,
	24*time.Hour,
)

var templateGroup singleflight.Group

func GetByURL(objectName string) (string, error) {
	if client == nil {
		return "", errors.New("MinIO client not initialized")
	}

	if cached, found := templateCache.Get(objectName); found {
		return cached.(string), nil
	}
	v, err, _ := templateGroup.Do(objectName, func() (interface{}, error) {

		if cached, found := templateCache.Get(objectName); found {
			return cached.(string), nil
		}

		object, err := client.GetObject(context.Background(), getBucket(), objectName, minio.GetObjectOptions{})
		if err != nil {
			return "", err
		}
		defer object.Close()

		content, err := io.ReadAll(object)
		if err != nil {
			return "", err
		}

		html := string(content)

		templateCache.SetDefault(objectName, html)

		return html, nil
	})

	if err != nil {
		return "", err
	}

	return v.(string), nil
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
