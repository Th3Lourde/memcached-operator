package controller

import (
	"crypto/rand"
	"fmt"
	v2 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
const length = 16
const admin = "admin"

// GenerateSecurePassword generates a random password to be used in a k8s secret
func GenerateSecurePassword() []byte {

	password := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))
	for i := range password {
		index, _ := rand.Int(rand.Reader, charsetLength)
		password[i] = charset[index.Int64()]
	}
	return password
}

func createPasswordSecret(clusterName, clusterNamespace string) *v2.Secret {
	secret := createSecret(clusterName, clusterNamespace, v2.SecretTypeBasicAuth)

	secret.Data = map[string][]byte{
		v2.BasicAuthUsernameKey: []byte(admin),
		v2.BasicAuthPasswordKey: GenerateSecurePassword(),
	}
	return secret
}

func createSecret(clusterName, clusterNamespace string, secretType v2.SecretType) *v2.Secret {
	return &v2.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-password", clusterName),
			Namespace: clusterNamespace,
		},
		Type: secretType,
		Data: map[string][]byte{},
	}
}
