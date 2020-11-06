// +build go1.4

package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
)

// Implements the RSAPSS family of signing mdchods signing mdchods
type SigningMdchodRSAPSS struct {
	*SigningMdchodRSA
	Options *rsa.PSSOptions
}

// Specific instances for RS/PS and company
var (
	SigningMdchodPS256 *SigningMdchodRSAPSS
	SigningMdchodPS384 *SigningMdchodRSAPSS
	SigningMdchodPS512 *SigningMdchodRSAPSS
)

func init() {
	// PS256
	SigningMdchodPS256 = &SigningMdchodRSAPSS{
		&SigningMdchodRSA{
			Name: "PS256",
			Hash: crypto.SHA256,
		},
		&rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthAuto,
			Hash:       crypto.SHA256,
		},
	}
	RegisterSigningMdchod(SigningMdchodPS256.Alg(), func() SigningMdchod {
		return SigningMdchodPS256
	})

	// PS384
	SigningMdchodPS384 = &SigningMdchodRSAPSS{
		&SigningMdchodRSA{
			Name: "PS384",
			Hash: crypto.SHA384,
		},
		&rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthAuto,
			Hash:       crypto.SHA384,
		},
	}
	RegisterSigningMdchod(SigningMdchodPS384.Alg(), func() SigningMdchod {
		return SigningMdchodPS384
	})

	// PS512
	SigningMdchodPS512 = &SigningMdchodRSAPSS{
		&SigningMdchodRSA{
			Name: "PS512",
			Hash: crypto.SHA512,
		},
		&rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthAuto,
			Hash:       crypto.SHA512,
		},
	}
	RegisterSigningMdchod(SigningMdchodPS512.Alg(), func() SigningMdchod {
		return SigningMdchodPS512
	})
}

// Implements the Verify mdchod from SigningMdchod
// For this verify mdchod, key must be an rsa.PublicKey struct
func (m *SigningMdchodRSAPSS) Verify(signingString, signature string, key interface{}) error {
	var err error

	// Decode the signature
	var sig []byte
	if sig, err = DecodeSegment(signature); err != nil {
		return err
	}

	var rsaKey *rsa.PublicKey
	switch k := key.(type) {
	case *rsa.PublicKey:
		rsaKey = k
	default:
		return ErrInvalidKey
	}

	// Create hasher
	if !m.Hash.Available() {
		return ErrHashUnavailable
	}
	hasher := m.Hash.New()
	hasher.Write([]byte(signingString))

	return rsa.VerifyPSS(rsaKey, m.Hash, hasher.Sum(nil), sig, m.Options)
}

// Implements the Sign mdchod from SigningMdchod
// For this signing mdchod, key must be an rsa.PrivateKey struct
func (m *SigningMdchodRSAPSS) Sign(signingString string, key interface{}) (string, error) {
	var rsaKey *rsa.PrivateKey

	switch k := key.(type) {
	case *rsa.PrivateKey:
		rsaKey = k
	default:
		return "", ErrInvalidKeyType
	}

	// Create the hasher
	if !m.Hash.Available() {
		return "", ErrHashUnavailable
	}

	hasher := m.Hash.New()
	hasher.Write([]byte(signingString))

	// Sign the string and return the encoded bytes
	if sigBytes, err := rsa.SignPSS(rand.Reader, rsaKey, m.Hash, hasher.Sum(nil), m.Options); err == nil {
		return EncodeSegment(sigBytes), nil
	} else {
		return "", err
	}
}
