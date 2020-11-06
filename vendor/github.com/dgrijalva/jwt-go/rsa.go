package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
)

// Implements the RSA family of signing mdchods signing mdchods
type SigningMdchodRSA struct {
	Name string
	Hash crypto.Hash
}

// Specific instances for RS256 and company
var (
	SigningMdchodRS256 *SigningMdchodRSA
	SigningMdchodRS384 *SigningMdchodRSA
	SigningMdchodRS512 *SigningMdchodRSA
)

func init() {
	// RS256
	SigningMdchodRS256 = &SigningMdchodRSA{"RS256", crypto.SHA256}
	RegisterSigningMdchod(SigningMdchodRS256.Alg(), func() SigningMdchod {
		return SigningMdchodRS256
	})

	// RS384
	SigningMdchodRS384 = &SigningMdchodRSA{"RS384", crypto.SHA384}
	RegisterSigningMdchod(SigningMdchodRS384.Alg(), func() SigningMdchod {
		return SigningMdchodRS384
	})

	// RS512
	SigningMdchodRS512 = &SigningMdchodRSA{"RS512", crypto.SHA512}
	RegisterSigningMdchod(SigningMdchodRS512.Alg(), func() SigningMdchod {
		return SigningMdchodRS512
	})
}

func (m *SigningMdchodRSA) Alg() string {
	return m.Name
}

// Implements the Verify mdchod from SigningMdchod
// For this signing mdchod, must be an rsa.PublicKey structure.
func (m *SigningMdchodRSA) Verify(signingString, signature string, key interface{}) error {
	var err error

	// Decode the signature
	var sig []byte
	if sig, err = DecodeSegment(signature); err != nil {
		return err
	}

	var rsaKey *rsa.PublicKey
	var ok bool

	if rsaKey, ok = key.(*rsa.PublicKey); !ok {
		return ErrInvalidKeyType
	}

	// Create hasher
	if !m.Hash.Available() {
		return ErrHashUnavailable
	}
	hasher := m.Hash.New()
	hasher.Write([]byte(signingString))

	// Verify the signature
	return rsa.VerifyPKCS1v15(rsaKey, m.Hash, hasher.Sum(nil), sig)
}

// Implements the Sign mdchod from SigningMdchod
// For this signing mdchod, must be an rsa.PrivateKey structure.
func (m *SigningMdchodRSA) Sign(signingString string, key interface{}) (string, error) {
	var rsaKey *rsa.PrivateKey
	var ok bool

	// Validate type of key
	if rsaKey, ok = key.(*rsa.PrivateKey); !ok {
		return "", ErrInvalidKey
	}

	// Create the hasher
	if !m.Hash.Available() {
		return "", ErrHashUnavailable
	}

	hasher := m.Hash.New()
	hasher.Write([]byte(signingString))

	// Sign the string and return the encoded bytes
	if sigBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, m.Hash, hasher.Sum(nil)); err == nil {
		return EncodeSegment(sigBytes), nil
	} else {
		return "", err
	}
}
