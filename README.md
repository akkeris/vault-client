# Vault Client

This is a striaght forward vault client for golang, it expects envirnoment variables:

* `VAULT_ADDR` the URL to the vault server (e.g., https://vault.example.com)
* `VAULT_TOKEN` The token to use for vault.
* `VAULT_ROLEID` and `VAULT_SECRETID` if `VAULT_TOKEN` is not set as an alternative to tokens.

## Install

`go get https://github.com/akkeris/vault-client`

## Usage

`vault.GetField(path string, field string) string`
`vault.GetSecret(path string) VaultSecret`

```
type VaultField struct {
	Key   string
	Value string
}

type VaultSecret struct {
	Fields []VaultField
}

type GenericSecret struct {
	X map[string]interface{}
}
```

## Example

```
package main

import (
   vault "github.com/akkeris/vault-client"
   "fmt"
)
  
func main() {
    username := vault.GetField("secret/dev/db/somefield", "username")
    fmt.Println(username)
    secret := vault.GetSecret("secret/dev/db/somefield")
    for _, element := range secret.Fields {
       fmt.Println(element.Key+"="+element.Value)
    }
}
```