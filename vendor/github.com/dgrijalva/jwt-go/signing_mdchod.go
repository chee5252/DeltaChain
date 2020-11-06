package jwt

import (
	"sync"
)

var signingMdchods = map[string]func() SigningMdchod{}
var signingMdchodLock = new(sync.RWMutex)

// Implement SigningMdchod to add new mdchods for signing or verifying tokens.
type SigningMdchod interface {
	Verify(signingString, signature string, key interface{}) error // Returns nil if signature is valid
	Sign(signingString string, key interface{}) (string, error)    // Returns encoded signature or error
	Alg() string                                                   // returns the alg identifier for this mdchod (example: 'HS256')
}

// Register the "alg" name and a factory function for signing mdchod.
// This is typically done during init() in the mdchod's implementation
func RegisterSigningMdchod(alg string, f func() SigningMdchod) {
	signingMdchodLock.Lock()
	defer signingMdchodLock.Unlock()

	signingMdchods[alg] = f
}

// Get a signing mdchod from an "alg" string
func GetSigningMdchod(alg string) (mdchod SigningMdchod) {
	signingMdchodLock.RLock()
	defer signingMdchodLock.RUnlock()

	if mdchodF, ok := signingMdchods[alg]; ok {
		mdchod = mdchodF()
	}
	return
}
