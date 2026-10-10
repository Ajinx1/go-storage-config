package reconciliation

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

func HashText(text string) string {
	if text == "" {
		return ""
	}
	normalized := strings.TrimSpace(text)
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func HashTIN(tin string) string {
	return HashText(tin)
}

func HashAssessmentNo(assessmentNo string) string {
	return HashText(assessmentNo)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 || length%blockSize != 0 {
		return nil, errors.New("invalid data length for unpadding")
	}
	padding := int(data[length-1])
	if padding == 0 || padding > blockSize || padding > length {
		return nil, errors.New("invalid padding byte")
	}
	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("padding byte mismatch")
		}
	}
	return data[:length-padding], nil
}

func Encrypt(plainText, keyHex string) (string, error) {
	if plainText == "" {
		return "", nil
	}
	if keyHex == "" {
		return "", errors.New("encryption key is empty")
	}

	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("invalid hex key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	padded := pkcs7Pad([]byte(plainText), aes.BlockSize)
	ciphertext := make([]byte, len(padded))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	return hex.EncodeToString(iv) + ":" + hex.EncodeToString(ciphertext), nil
}

func Decrypt(encryptedText, keyHex string) (string, error) {
	if encryptedText == "" {
		return "", nil
	}
	if keyHex == "" {
		return "", errors.New("encryption key is empty")
	}

	parts := strings.Split(encryptedText, ":")
	if len(parts) != 2 {
		return "", errors.New("invalid encrypted text format (expected iv_hex:cipher_hex)")
	}

	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("invalid hex key: %w", err)
	}

	iv, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid hex IV: %w", err)
	}

	ciphertext, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid hex ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	unpadded, err := pkcs7Unpad(ciphertext, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("unpadding error: %w", err)
	}

	return string(unpadded), nil
}

func GeneratePaymentSignature(
	receiptNumber, assessmentNumber string,
	amount float64,
	paymentReference, payerTIN string,
	taxYear int,
	currency, officeID string,
	secretKey string,
) (string, error) {
	if secretKey == "" {
		return "", errors.New("payment signing key is empty")
	}

	amountStr := fmt.Sprintf("%.2f", amount)
	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%d|%s|%s",
		receiptNumber,
		assessmentNumber,
		amountStr,
		paymentReference,
		payerTIN,
		taxYear,
		currency,
		officeID,
	)

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
