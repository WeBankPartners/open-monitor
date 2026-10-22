package k8s

import (
	"errors"
	"github.com/WeBankPartners/go-common-lib/cipher"
	"strings"
)

type Credentials struct {
	Guid           string `json:"-"`
	APIServer      string `json:"-"`
	CA             string `json:"-"`
	EncryptedToken string `json:"-"`
}

// Resolve returns the effective cluster connection settings. Namespace
// registrations may carry only an optional token override; API server and CA
// always come from the cluster registration.
func Resolve(namespace, cluster Credentials) Credentials {
	result := cluster
	if strings.TrimSpace(namespace.EncryptedToken) != "" {
		result.Guid = namespace.Guid
		result.EncryptedToken = namespace.EncryptedToken
	}
	return result
}

func PrepareCredentials(guid, seed, endpoint, ca, token string) (Credentials, error) {
	if strings.TrimSpace(guid) == "" || seed == "" || strings.TrimSpace(token) == "" {
		return Credentials{}, errors.New("guid, encryption seed and credential are required")
	}
	plain := token
	for _, prefix := range cipher.CIPHER_MAP {
		if strings.HasPrefix(token, prefix) {
			var err error
			plain, err = cipher.AesDePasswordByGuid(guid, seed, token)
			if err != nil {
				return Credentials{}, errors.New("cannot decrypt platform credential")
			}
			break
		}
	}
	encrypted, err := cipher.AesEnPasswordByGuid(guid, seed, plain, "")
	if err != nil {
		return Credentials{}, errors.New("cannot encrypt namespace credential")
	}
	return Credentials{Guid: guid, APIServer: endpoint, CA: ca, EncryptedToken: encrypted}, nil
}

func PrepareOptionalToken(guid, seed, token string) (Credentials, error) {
	if strings.TrimSpace(guid) == "" || seed == "" {
		return Credentials{}, errors.New("guid and encryption seed are required")
	}
	if strings.TrimSpace(token) == "" {
		return Credentials{Guid: guid}, nil
	}
	return PrepareCredentials(guid, seed, "", "", token)
}
func (c Credentials) PlainToken(seed string) (string, error) {
	plain, err := cipher.AesDePasswordByGuid(c.Guid, seed, c.EncryptedToken)
	if err != nil {
		return "", errors.New("cannot decrypt namespace credential")
	}
	return plain, nil
}
