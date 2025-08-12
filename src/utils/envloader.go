package utils

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("⚠️  Unable to detect runtime path for .env loading")
		return
	}

	rootPath := filepath.Join(filepath.Dir(b), "../..", ".env")

	err := godotenv.Load(rootPath)
	if err != nil {
		log.Printf("⚠️  No .env file found at %s\n", rootPath)
	}
}
