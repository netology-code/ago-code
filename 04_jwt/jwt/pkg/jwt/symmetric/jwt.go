package symmetric

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// зашили для простоты
var defaultHeader = Header{
	Alg: "HS256",
	Typ: "JWT",
}

func Encode(payload interface{}, key []byte) (token string, err error) {
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

	signatureEncoded, err := calculateSignatureEncoded(headerEncoded, payloadEncoded, key)
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

func Verify(token string, key []byte) (ok bool, err error) {
	parts, err := splitToken(token)
	if err != nil {
		return false, err
	}
	headerEncoded, payloadEncoded, signatureEncoded := parts[0], parts[1], parts[2]

	verificationEncoded, err := calculateSignatureEncoded(headerEncoded, payloadEncoded, key)
	if err != nil {
		return false, nil
	}
	return signatureEncoded == verificationEncoded, nil
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
	h := hmac.New(sha256.New, key)
	h.Write([]byte(headerEncoded + "." + payloadEncoded))
	signature := h.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(signature), nil
}
