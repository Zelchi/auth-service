package jwt

import (
	"errors"
	"os"
	"strings"
)

const minSecretBytes = 32

func validateSecret(name, value string) ([]byte, error) {
	secret := strings.TrimSpace(value)
	if len([]byte(secret)) < minSecretBytes {
		return nil, errors.New(name + " deve ter pelo menos 32 bytes")
	}
	trimmed := strings.ToLower(secret)
	if strings.Contains(trimmed, "replace-with") || strings.Contains(trimmed, "change-me") ||
		strings.Contains(trimmed, "gere-um-segredo") {
		return nil, errors.New(name + " de exemplo não pode ser usado")
	}
	return []byte(secret), nil
}

func signingKeys() ([][]byte, error) {
	active, err := validateSecret("JWT_SECRET", os.Getenv("JWT_SECRET"))
	if err != nil {
		return nil, err
	}
	keys := [][]byte{active}
	if previous := strings.TrimSpace(os.Getenv("JWT_SECRET_PREVIOUS")); previous != "" {
		previousKey, err := validateSecret("JWT_SECRET_PREVIOUS", previous)
		if err != nil {
			return nil, err
		}
		keys = append(keys, previousKey)
	}
	return keys, nil
}

func signingKey() ([]byte, error) {
	keys, err := signingKeys()
	if err != nil {
		return nil, err
	}

	return keys[0], nil
}

// ValidateSecret permite validar a configuração no início da aplicação sem
// expor o segredo ou duplicar as regras no pacote main.
func ValidateSecret() error {
	_, err := signingKeys()
	return err
}
