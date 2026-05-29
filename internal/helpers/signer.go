package helpers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

type Signer struct {
	key    string
	header string
}

func NewSigner(key string) *Signer {
	return &Signer{key: key, header: "HashSHA256"}
}

func (s *Signer) getSign(data []byte) string {
	h := hmac.New(sha256.New, []byte(s.key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Signer) SignRequest(data []byte, request *http.Request) {
	if s.key == ""{ return}
	sign := s.getSign(data)
	request.Header.Set(s.header, sign)
}
func (s *Signer) SignResponse(data []byte, response http.ResponseWriter) {
	if s.key == ""{ return}
	sign := s.getSign(data)
	response.Header().Set(s.header, sign)
}

func (s *Signer) VerifyRequest(req *http.Request) bool {
	if s.key == "" {
		return true
	}
	origin := req.Header.Get(s.header)
	if origin == "" {
		return true
	} else {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			return false
		}
		calculatedSign := s.getSign(data)
		src, err := hex.DecodeString(origin)
		if err != nil {
			return false
		}
		dst, err := hex.DecodeString(calculatedSign)
		if err != nil {
			return false
		}
		req.Body = io.NopCloser(bytes.NewReader(data))
		return hmac.Equal(src, dst)
	}
}
