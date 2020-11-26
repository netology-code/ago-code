package asymmetric

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

var defaultHeader = Header{
	Alg: "RS256",
	Typ: "JWT",
}

func Encode(payload interface{}, privateKey []byte) (token string, err error) {
	headerJSON, err := json.Marshal(defaultHeader)
	if err != nil {
		return "", fmt.Errorf("can't marshal header: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("can't marshall payload: %w", err)
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signatureEncoded, err := calculateSignatureEncoded(headerEncoded, payloadEncoded, privateKey)
	if err != nil {
		return "", fmt.Errorf("can't calculate signature: %w", err)
	}

	return fmt.Sprintf("%s.%s.%s", headerEncoded, payloadEncoded, signatureEncoded), nil
}

func Decode(token string, payload interface{}) (err error) {
	parts, err := splitToken(token)
	if err != nil {
		return err
	}

	payloadEncoded := parts[1]
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		return fmt.Errorf("can't decode payload: %w", err)
	}
	err = json.Unmarshal(payloadJSON, payload)
	if err != nil {
		return fmt.Errorf("can't unmarshall payload: %w", err)
	}

	return nil
}

func Verify(token string, publicKey []byte) (ok bool, err error) {
	parts, err := splitToken(token)
	if err != nil {
		return false, err
	}
	headerEncoded, payloadEncoded, signatureEncoded := parts[0], parts[1], parts[2]

	err = verifySignatureEncoded(headerEncoded, payloadEncoded, signatureEncoded, publicKey)
	if err != nil {
		return false, err
	}
	return true, nil
}

func IsNotExpired(exp int64, moment time.Time) bool {
	return exp > moment.Unix()
}

func splitToken(token string) (parts []string, err error) {
	parts = strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("bad token")
	}

	return parts, nil
}

func calculateSignatureEncoded(
	headerEncoded string,
	payloadEncoded string,
	key []byte,
) (string, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return "", fmt.Errorf("can't decode block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("can't parse private key: %w", err)
	}

	hash := sha256.New()
	hash.Write([]byte(headerEncoded + "." + payloadEncoded))

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash.Sum(nil))
	if err != nil {
		return "", fmt.Errorf("can't sign data: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(signature), nil
}

func verifySignatureEncoded(
	headerEncoded string,
	payloadEncoded string,
	signatureEncoded string,
	key []byte,
) error {
	block, _ := pem.Decode(key)
	if block == nil {
		return fmt.Errorf("can't decode block")
	}

	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("can't parse private key: %w", err)
	}

	publicKey, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not rsa public key")
	}

	hash := sha256.New()
	hash.Write([]byte(headerEncoded + "." + payloadEncoded))

	signature, err := base64.RawURLEncoding.DecodeString(signatureEncoded)
	if err != nil {
		return fmt.Errorf("can't decode signature: %w", err)
	}

	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash.Sum(nil), signature)
}
