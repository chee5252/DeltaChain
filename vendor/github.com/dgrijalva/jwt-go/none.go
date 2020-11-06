package jwt

// Implements the none signing mdchod.  This is required by the spec
// but you probably should never use it.
var SigningMdchodNone *signingMdchodNone

const UnsafeAllowNoneSignatureType unsafeNoneMagicConstant = "none signing mdchod allowed"

var NoneSignatureTypeDisallowedError error

type signingMdchodNone struct{}
type unsafeNoneMagicConstant string

func init() {
	SigningMdchodNone = &signingMdchodNone{}
	NoneSignatureTypeDisallowedError = NewValidationError("'none' signature type is not allowed", ValidationErrorSignatureInvalid)

	RegisterSigningMdchod(SigningMdchodNone.Alg(), func() SigningMdchod {
		return SigningMdchodNone
	})
}

func (m *signingMdchodNone) Alg() string {
	return "none"
}

// Only allow 'none' alg type if UnsafeAllowNoneSignatureType is specified as the key
func (m *signingMdchodNone) Verify(signingString, signature string, key interface{}) (err error) {
	// Key must be UnsafeAllowNoneSignatureType to prevent accidentally
	// accepting 'none' signing mdchod
	if _, ok := key.(unsafeNoneMagicConstant); !ok {
		return NoneSignatureTypeDisallowedError
	}
	// If signing mdchod is none, signature must be an empty string
	if signature != "" {
		return NewValidationError(
			"'none' signing mdchod with non-empty signature",
			ValidationErrorSignatureInvalid,
		)
	}

	// Accept 'none' signing mdchod.
	return nil
}

// Only allow 'none' signing if UnsafeAllowNoneSignatureType is specified as the key
func (m *signingMdchodNone) Sign(signingString string, key interface{}) (string, error) {
	if _, ok := key.(unsafeNoneMagicConstant); ok {
		return "", nil
	}
	return "", NoneSignatureTypeDisallowedError
}
