package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchSecrets_MatchByKey(t *testing.T) {
	client := listClient(t, map[string]interface{}{
		"data": map[string]interface{}{
			"keys": []interface{}{"db/password"},
		},
	})

	client.logical.(*fakeLogical).responses["secret/data/db/password"] = &fakeSecret{
		data: map[string]interface{}{
			"data": map[string]interface{}{
				"db_password": "supersecret",
				"host":        "localhost",
			},
		},
	}

	results, err := SearchSecrets(client, "secret", "password")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "db/password", results[0].Path)
	assert.Contains(t, results[0].MatchedKeys, "db_password")
}

func TestSearchSecrets_MatchByValue(t *testing.T) {
	client := listClient(t, map[string]interface{}{
		"data": map[string]interface{}{
			"keys": []interface{}{"app/config"},
		},
	})

	client.logical.(*fakeLogical).responses["secret/data/app/config"] = &fakeSecret{
		data: map[string]interface{}{
			"data": map[string]interface{}{
				"api_key": "FIND_ME_123",
				"region":  "us-east-1",
			},
		},
	}

	results, err := SearchSecrets(client, "secret", "find_me")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].MatchedKeys, "api_key")
}

func TestSearchSecrets_NoMatch(t *testing.T) {
	client := listClient(t, map[string]interface{}{
		"data": map[string]interface{}{
			"keys": []interface{}{"app/config"},
		},
	})

	client.logical.(*fakeLogical).responses["secret/data/app/config"] = &fakeSecret{
		data: map[string]interface{}{
			"data": map[string]interface{}{
				"region": "us-east-1",
			},
		},
	}

	results, err := SearchSecrets(client, "secret", "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, results)
}
