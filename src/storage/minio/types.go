package minio

import "io"

type UploadInput struct {
	File        io.Reader
	Size        int64
	ObjectName  string
	ContentType string
}

type DeleteInput struct {
	ObjectName string
}

type ListInput struct {
	Prefix string
}
