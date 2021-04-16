package formatters

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
)

// CreateRandomString ...
func CreateRandomString() string {
	return fmt.Sprintf("rand:%d", rand.Int()) // #nosec G404
}

// ConvertFormattedPrivateKey into secp256k1r private key type
func ConvertFormattedPrivateKey(pkStr string) (*crypto.PrivateKeySECP256K1R, error) {
	factory := crypto.FactorySECP256K1R{}
	if !strings.HasPrefix(pkStr, constants.SecretKeyPrefix) {
		return nil, fmt.Errorf("private key missing %s prefix", constants.SecretKeyPrefix)
	}
	trimmedPrivateKey := strings.TrimPrefix(pkStr, constants.SecretKeyPrefix)
	pkBytes, err := formatting.Decode(formatting.CB58, trimmedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("problem parsing private key: %w", err)
	}

	skIntf, err := factory.ToPrivateKey(pkBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key bytes to private key due to: %w", err)
	}
	sk := skIntf.(*crypto.PrivateKeySECP256K1R)
	return sk, nil
}
