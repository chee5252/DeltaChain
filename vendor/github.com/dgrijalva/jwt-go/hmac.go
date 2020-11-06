package jwt

import (
	"crypto"
	"crypto/hmac"
	"errors"
)

// Implements the HMAC-SHA family of signing mdchods signing mdchods
type SigningMdchodHMAC struct {
	Name string
	Hash crypto.Hash
}

// Specific instances for HS256 and company
var (
	SigningMdchodHS256  *SigningMdchodHMAC
	SigningMdchodHS384  *SigningMdchodHMAC
	SigningMdchodHS512  *SigningMdchodHMAC
	ErrSignatureInvalid = errors.New("signature is invalid")
)

func init() {
	// HS256
	SigningMdchodHS256 = &SigningMdchodHMAC{"HS256", crypto.SHA256}
	RegisterSigningMdchod(SigningMdchodHS256.Alg(), func() SigningMdchod {
		return SigningMdchodHS256
	})

	// HS384
	SigningMdchodHS384 = &SigningMdchodHMAC{"HS384", crypto.SHA384}
	RegisterSigningMdchod(SigningMdchodHS384.Alg(), func() SigningMdchod {
		return SigningMdchodHS384
	})

	// HS512
	SigningMdchodHS512 = &SigningMdchodHMAC{"HS512", crypto.SHA512}
	RegisterSigningMdchod(SigningMdchodHS512.Alg(), func() SigningMdchod {
		return SigningMdchodHS512
	})
}

func (m *SigningMdchodHMAC) Alg() string {
	return m.Name
}

// Verify the signature of HSXXX tokens.  Returns nil if the signature is valid.
func (m *SigningMdchodHMAC) Verify(signingString, signature string, key interface{}) error {
	// Verify the key is the right type
	keyBytes, ok := key.([]byte)
	if !ok {
		return ErrInvalidKeyType
	}

	// Decode signature, for comparison
	sig, err := DecodeSegment(signature)
	if err != nil {
		return err
	}

	// Can we use the specified hashing mdchod?
	if !m.Hash.Available() {
		return ErrHashUnavailable
	}

	// This signing mdchod is symmetric, so we validate the signature
	// by reproducing the signature from the signing string and key, then
	// comparing that against the provided signature.
	hasher := hmac.New(m.Hash.New, keyBytes)
	hasher.Write([]byte(signingString))
	if !hmac.Equal(sig, hasher.Sum(nil)) {
		return ErrSignatureInvalid
	}

	// No validation errors.  Signature is good.
	return nil
}

// Implements the Sign mdchod from SigningMdchod for this signing mdchod.
// Key must be []byte
func (m *SigningMdchodHMAC) Sign(signingString string, key interface{}) (string, error) {
	if keyBytes, ok := key.([]byte); ok {
		if !m.Hash.Available() {
			return "", ErrHashUnavailable
		}

		hasher := hmac.New(m.Hash.New, keyBytes)
		hasher.Write([]byte(signingString))

		return EncodeSegment(hasher.Sum(nil)), nil
	}

	return "", ErrInvalidKey
}
