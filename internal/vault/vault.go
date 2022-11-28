package vault

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/vault/api"
)

// Provider ...
//type Provider struct {
//	path    string
//	client  *api.Logical
//	results map[string]map[string]string
//}

type vaultClient struct {
	client *api.Client
	secret *api.Secret
}

type pinger interface {
	ping() error
}

type dbPinger struct {
	db *sql.DB
}

func (d dbPinger) ping() error {
	return d.db.Ping()
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

func NewVaultClient(address string) (*vaultClient, error) {
	config := api.Config{Address: address}
	client, err := api.NewClient(&config)
	if err != nil {
		return nil, err
	}
	return &vaultClient{client: client}, nil
}

func (v *vaultClient) readVaultSecret(path string) error {
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return err
	}
	v.secret = secret
	return nil
}

func (v *vaultClient) GetCredentials(vaultAddr, vaultToken, vaultRoleID, vaultSecretID string) (string, string, error) {
	if vaultRoleID != "" && vaultSecretID != "" {
		options := map[string]interface{}{
			"role_id":   vaultRoleID,
			"secret_id": vaultSecretID,
		}
		path := "auth/approle/login"
		pingExternalService(vaultAddr, &vaultAppRolePinger{v.client, path, options})
		secret, err := v.client.Logical().Write(path, options)
		if err != nil {
			return "", "", err
		}
		v.client.SetToken(secret.Auth.ClientToken)
		if secret.Auth == nil {
			return "", "", fmt.Errorf("could not read auth info from secret")
		}
		err = v.readVaultSecret("database/creds/vault-mysql-role")
		if err != nil {
			return "", "", err
		}
		username := v.secret.Data["username"].(string)
		password := v.secret.Data["password"].(string)
		return username, password, nil
	}
	if vaultToken != "" {
		v.client.SetToken(vaultToken)
		path := "secrets/database"
		pingExternalService(vaultAddr, &vaultPinger{v.client, path})
		err := v.readVaultSecret(path)
		if err != nil {
			return "", "", err
		}
		username := v.secret.Data["username"].(string) //???
		password := v.secret.Data["password"].(string)
		return username, password, nil
	}
	return "", "", fmt.Errorf("could not read vault secret")
}

func (v *vaultClient) renewLease() error {
	log.Printf("Renewing lease %v.", v.secret.LeaseID)
	_, err := v.client.Sys().Renew(v.secret.LeaseID, v.secret.LeaseDuration)
	if err != nil {
		return err
	}
	return nil
}

func (v *vaultClient) RegularlyRenewLease() error {
	if !v.secret.Renewable {
		log.Println("Cowardly refusing to renew unrenewable secret.")
		return nil
	}
	v.renewLease()
	// renew lease 100 seconds before expiry
	interval := time.Duration(v.secret.LeaseDuration)*time.Second - 100*time.Second
	renewTicker := time.NewTicker(interval)
	log.Printf("Scheduling regular renewal for lease %s every %v", v.secret.LeaseID, interval)

	for {
		select {
		case <-renewTicker.C:
			v.renewLease()
		}
	}
}

// pingExternalService pings an external service with linearly increasing backoff time.
func pingExternalService(addr string, pinger pinger) error {
	numBackOffIterations := 15
	for i := 1; i <= numBackOffIterations; i++ {
		log.Printf("Pinging %s.\n", addr)
		err := pinger.ping()
		if err != nil {
			log.Println(err)
		}
		if err == nil {
			log.Printf("Connected to %s.", addr)
			break
		}
		waitDuration := time.Duration(i) * time.Second
		log.Printf("Backing off for %v.\n", waitDuration)
		time.Sleep(waitDuration)
		if i == numBackOffIterations {
			return err
		}
	}
	return nil
}

type vaultPinger struct {
	vaultClient *api.Client
	path        string
}

func (v *vaultPinger) ping() error {
	_, err := v.vaultClient.Logical().Read(v.path)
	return err
}

type vaultAppRolePinger struct {
	vaultClient *api.Client
	path        string
	options     map[string]interface{}
}

func (v *vaultAppRolePinger) ping() error {
	_, err := v.vaultClient.Logical().Write(v.path, v.options)
	return err
}

//// New ...
//func New() (*Provider, error) {
//
//	token := env.ViperEnvVariable("VAULT_TOKEN")
//	addr := env.ViperEnvVariable("VAULT_ADDRESS")
//	path := env.ViperEnvVariable("VAULT_PATH")
//
//	config := &vault.Config{
//		Address: addr,
//	}
//
//	client, err := vault.NewClient(config)
//	if err != nil {
//		return nil, fmt.Errorf("new client: %w", err)
//	}
//
//	client.SetToken(token)
//
//	return &Provider{
//		path:    path,
//		client:  client.Logical(),
//		results: make(map[string]map[string]string),
//	}, nil
//}

//// Get retrieves a value from vault using the KV engine. The actual key selected is determined by the value
//// separated by the colon. For example "database:password" will retrieve the key "password" from the path
//// "database".
//func (p *Provider) Get(v string) (string, error) {
//	// <path>/data/<path-secret>:key
//	split := strings.Split(v, ":")
//	if len(split) == 1 {
//		return "", errors.New("missing key value")
//	}
//
//	pathSecret := split[0]
//	key := split[1]
//
//	res, ok := p.results[pathSecret]
//	if ok {
//		val, ok := res[key]
//		if !ok {
//			return "", errors.New("key not found in cached data")
//		}
//
//		return val, nil
//	}
//
//	secret, err := p.client.Read(fmt.Sprintf("%s/data/%s", p.path, pathSecret))
//	if err != nil {
//		return "", fmt.Errorf("reading: %w", err)
//	}
//
//	if secret == nil {
//		return "", errors.New("secret not found")
//	}
//
//	data, ok := secret.Data["data"].(map[string]interface{})
//	if !ok {
//		return "", errors.New("invalid data in secret")
//	}
//
//	secrets := make(map[string]string)
//
//	for k, v := range data {
//		val, ok := v.(string)
//		if !ok {
//			return "", errors.New("secret value in data is not string")
//		}
//
//		secrets[k] = val
//	}
//
//	val, ok := secrets[key]
//	if !ok {
//		return "", errors.New("key not found in retrieved data")
//	}
//
//	p.results[pathSecret] = secrets
//
//	return val, nil
//}
