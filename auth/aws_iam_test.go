package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerManagerForAwsIAM(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		accountId  string
		allowUsers []string
		isErr      bool
	}{
		"fail": {isErr: true},
		"success": {
			accountId:  "testid",
			allowUsers: []string{"gjbae1212", "test"},
		},
	}

	for _, t := range tests {
		s, err := NewServerManagerForAwsIAM(t.accountId, t.allowUsers)
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			assert.Equal(t.accountId, s.(*AwsIamConfig).ServerAccountId)
			assert.Equal(t.allowUsers, s.(*AwsIamConfig).ServerAllowUsers)
		}
	}

}

func TestNewClientManagerForAwsIAM(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		accessKey    string
		accessSecret string
		isErr        bool
	}{
		"fail": {isErr: true},
		"success": {
			accessKey:    "test-key",
			accessSecret: "test-secret",
		},
	}

	for _, t := range tests {
		s, err := NewClientManagerForAwsIAM(t.accessKey, t.accessSecret)
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			assert.Equal(t.accessKey, s.(*AwsIamConfig).ClientAccessKey)
			assert.Equal(t.accessSecret, s.(*AwsIamConfig).ClientSecretAccessKey)
		}
	}
}

func TestAwsIamConfig_ServerAuth(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		accountId string
		ok        bool
	}{
		"success": {
			accountId: "test-id",
			ok:        true,
		},
	}

	for _, t := range tests {
		s, err := NewServerManagerForAwsIAM(t.accountId, nil)
		assert.NoError(err)
		_, ok := s.ServerAuth()
		assert.Equal(t.ok, ok)
	}
}

func TestAwsIamConfig_ClientAuth(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		accessKey    string
		accessSecret string
		ok           bool
	}{
		"success": {
			accessKey:    "test-id",
			accessSecret: "test-secret",
			ok:           true,
		},
	}

	for _, t := range tests {
		s, err := NewClientManagerForAwsIAM(t.accessKey, t.accessSecret)
		assert.NoError(err)
		_, ok := s.ClientAuth()
		assert.Equal(t.ok, ok)
	}

}
