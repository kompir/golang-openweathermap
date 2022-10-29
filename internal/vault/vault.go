package vault

import (
	"errors"
	"fmt"
	env "github.com/kompir/golang-openweathermap/database"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

// Provider ...
type Provider struct {
	path    string
	client  *vault.Logical
	results map[string]map[string]string
}

//// NewVaultProvider instantiates the Vault client using configuration defined in environment variables.
//func NewVaultProvider() (*Provider, error) {
//	// XXX: We will revisit this code in future episodes replacing it with another solution
//	vaultPath := env.ViperEnvVariable("VAULT_PATH")
//	vaultToken := env.ViperEnvVariable("VAULT_TOKEN")
//	vaultAddress := env.ViperEnvVariable("VAULT_ADDRESS")
//	// XXX: -
//
//	provider, err := New(vaultToken, vaultAddress, vaultPath)
//	if err != nil {
//		return nil, fmt.Errorf("Vault new: %w", err)
//	}
//
//	return provider, nil
//}

// New ...
func New() (*Provider, error) {

	token := env.ViperEnvVariable("VAULT_TOKEN")
	addr := env.ViperEnvVariable("VAULT_ADDRESS")
	path := env.ViperEnvVariable("VAULT_PATH")

	config := &vault.Config{
		Address: addr,
	}

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	client.SetToken(token)

	return &Provider{
		path:    path,
		client:  client.Logical(),
		results: make(map[string]map[string]string),
	}, nil
}

// Get retrieves a value from vault using the KV engine. The actual key selected is determined by the value
// separated by the colon. For example "database:password" will retrieve the key "password" from the path
// "database".
func (p *Provider) Get(v string) (string, error) {
	// <path>/data/<path-secret>:key
	split := strings.Split(v, ":")
	if len(split) == 1 {
		return "", errors.New("missing key value")
	}

	pathSecret := split[0]
	key := split[1]

	res, ok := p.results[pathSecret]
	if ok {
		val, ok := res[key]
		if !ok {
			return "", errors.New("key not found in cached data")
		}

		return val, nil
	}

	secret, err := p.client.Read(fmt.Sprintf("%s/data/%s", p.path, pathSecret))
	if err != nil {
		return "", fmt.Errorf("reading: %w", err)
	}

	if secret == nil {
		return "", errors.New("secret not found")
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", errors.New("invalid data in secret")
	}

	secrets := make(map[string]string)

	for k, v := range data {
		val, ok := v.(string)
		if !ok {
			return "", errors.New("secret value in data is not string")
		}

		secrets[k] = val
	}

	val, ok := secrets[key]
	if !ok {
		return "", errors.New("key not found in retrieved data")
	}

	p.results[pathSecret] = secrets

	return val, nil
}
