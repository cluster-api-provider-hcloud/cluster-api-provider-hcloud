package scope

import (
	"context"
	"crypto/rand"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (s *ClusterScope) EnsureCiliumIPSecKeysExists(ctx context.Context) (keys *string, err error) {
	keyRef := s.HcloudCluster.Spec.CNI.Network.Cilium.IPSecKeysRef

	var keySecret corev1.Secret
	var secretName = types.NamespacedName{
		Namespace: s.Namespace(),
		Name:      keyRef.Name,
	}

	if err := s.Client.Get(ctx, secretName, &keySecret); err == nil {
		key, ok := keySecret.Data[keyRef.Key]
		if !ok {
			return nil, fmt.Errorf("no key '%s' found in secret '%s' for IPSecKeysRef", keyRef.Key, secretName)
		}
		keyString := string(key)
		return &keyString, nil
	} else if !apierrors.IsNotFound(err) {
		return nil, err
	}

	// ipsec keys not found and need to be recreated

	// generate 20 random bytes
	randomBytes := make([]byte, 20)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("error creating random bytes for IPSecKeysRef: %w", err)
	}
	keyID := 3
	keyString := fmt.Sprintf(
		"%d rfc4106(gcm(aes)) %x 128",
		keyID,
		randomBytes,
	)

	keySecret = corev1.Secret{}
	keySecret.Namespace = secretName.Namespace
	keySecret.Name = secretName.Name

	keySecret.Data = map[string][]byte{
		keyRef.Key: []byte(keyString),
	}

	if err := s.Client.Create(ctx, &keySecret); err != nil {
		return nil, fmt.Errorf("error creating new secret for IPSecKeysRef: %w", err)
	}

	return &keyString, nil
}
