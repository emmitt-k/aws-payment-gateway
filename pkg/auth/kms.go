package auth

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

// KMSClient provides encryption and decryption operations using AWS KMS
type KMSClient struct {
	client *kms.Client
	keyID  string
}

// NewKMSClient creates a new KMS client
func NewKMSClient(ctx context.Context, region, keyID string) (*KMSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := kms.NewFromConfig(cfg)

	return &KMSClient{
		client: client,
		keyID:  keyID,
	}, nil
}

// EncryptData encrypts data using KMS
func (k *KMSClient) EncryptData(ctx context.Context, plaintext []byte) ([]byte, error) {
	input := &kms.EncryptInput{
		KeyId:     aws.String(k.keyID),
		Plaintext: plaintext,
	}

	result, err := k.client.Encrypt(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	return result.CiphertextBlob, nil
}

// DecryptData decrypts data using KMS
func (k *KMSClient) DecryptData(ctx context.Context, ciphertext []byte) ([]byte, error) {
	input := &kms.DecryptInput{
		CiphertextBlob: ciphertext,
	}

	result, err := k.client.Decrypt(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return result.Plaintext, nil
}

// GenerateDataKey generates a data key using KMS
func (k *KMSClient) GenerateDataKey(ctx context.Context) ([]byte, []byte, error) {
	input := &kms.GenerateDataKeyInput{
		KeyId:   aws.String(k.keyID),
		KeySpec: types.DataKeySpecAes256,
	}

	result, err := k.client.GenerateDataKey(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate data key: %w", err)
	}

	return result.Plaintext, result.CiphertextBlob, nil
}
