package signup

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/kataras/jwt"
	"github.com/mantil-io/mantil.go"
)

type TokenClaims struct {
	ActivationCode string `json:"activationCode,omitempty"`
	ActivationID   string `json:"activationID,omitempty"`
	CreatedAt      int64  `json:"createdAt,omitempty"`
}

func (s *Signup) generateToken(claims interface{}) (string, error) {
	_, privateKey, err := s.keys()
	if err != nil {
		return "", err
	}
	buf, err := jwt.Sign(jwt.EdDSA, ed25519.PrivateKey(privateKey), claims, jwt.MaxAge(time.Hour*24*365))
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (s *Signup) decodeToken(token string, claims interface{}) error {
	publicKey, _, err := s.keys()
	if err != nil {
		return err
	}
	verifiedToken, err := jwt.Verify(jwt.EdDSA, ed25519.PublicKey(publicKey), []byte(token))
	if err != nil {
		return fmt.Errorf("token verify failed: %w", err)
	}

	return verifiedToken.Claims(&claims)
}

type keyPair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

const keyPairKVKey = "keys"

func (s *Signup) keys() (string, string, error) {
	var kp keyPair
	err := s.kv.Keys().Get(keyPairKVKey, &kp)
	if errors.As(err, &mantil.ErrItemNotFound{}) {
		public, private, err := generateKeyPair()
		if err != nil {
			return "", "", err
		}
		kp.Public = public
		kp.Private = private
		if err := s.kv.Keys().Put(keyPairKVKey, &kp); err != nil {
			return "", "", err
		}
	} else if err != nil {
		return "", "", err
	}
	public, err := jwt.Base64Decode([]byte(kp.Public))
	if err != nil {
		return "", "", fmt.Errorf("failed to decode key %w", err)
	}
	private, err := jwt.Base64Decode([]byte(kp.Private))
	if err != nil {
		return "", "", fmt.Errorf("failed to decode key %w", err)
	}
	return string(public), string(private), nil
}

func generateKeyPair() (string, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return string(jwt.Base64Encode(publicKey)), string(jwt.Base64Encode(privateKey)), nil
}
