package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type VaultField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type VaultSecret struct {
	Fields []VaultField
}

type GenericSecret struct {
	X map[string]interface{}
}

type VaultList struct {
	LeaseID       string `json:"lease_id"`
	Renewable     bool   `json:"renewable"`
	LeaseDuration int    `json:"lease_duration"`
	Data          struct {
		Keys []string `json:"keys"`
	} `json:"data"`
	Warnings interface{} `json:"warnings"`
	Auth     interface{} `json:"auth"`
}

func ListSecrets(path string) ([]string, error) {
	var list []string
	req, err := http.NewRequest("GET", path + "?list=true", nil)
	if err != nil {
		return list, err
	}
	req.Header.Add("X-Vault-Token", getToken())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return list, err
	}
	defer resp.Body.Close()
	if err != nil {
		return list, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return list, err
	}
	var vaultlist VaultList
	_ = json.Unmarshal(respBody, &vaultlist)
	if len(vaultlist.Data.Keys) == 0 {
		list = append(list, "vault:"+path)
	}
	for _, element := range vaultlist.Data.Keys {
		if strings.HasSuffix(element, "/") {
			newelement := strings.Replace(element, "/", "", -1)
			rlist, err := ListSecrets(path + "/" + newelement)
			if err != nil {
				return list, err
			}
			list = append(list, rlist...)
		}
		if !strings.HasSuffix(element, "/") {
			list = append(list, "vault:"+path+"/"+element)
		}
	}
	return list, nil
}

func WriteField(path string, key string, content string) (e error) {
	existingsecret := GetSecret(path)
	var newfield VaultField
	newfield.Key = key
	newfield.Value = content
	existingsecret.Fields = append(existingsecret.Fields, newfield)
	var newsecret GenericSecret
	var m map[string]interface{}
	m = make(map[string]interface{})

	for _, element := range existingsecret.Fields {
		m[element.Key] = element.Value
	}
	newsecret.X = m
	jsonStr, err := json.Marshal(newsecret.X)
	if err != nil {
		fmt.Println(err)
		return err
	}
	req, err := http.NewRequest("POST", getAddress()+"/v1/"+path, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Vault-Token", getToken())
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetFieldFromVaultSecret(secret VaultSecret, field string) string {
	var toreturn string
	for _, element := range secret.Fields {
		if element.Key == field {
			toreturn = element.Value
		}
	}
	return toreturn
}

func GetField(path string, field string) string {
	return GetFieldFromVaultSecret(GetSecret(path), field)
}

func GetSecret(path string) VaultSecret {
	vaultaddruri := getAddress() + "/v1/" + path
	vreq, err := http.NewRequest("GET", vaultaddruri, nil)
	vreq.Header.Add("X-Vault-Token", getToken())
	vclient := &http.Client{}
	vresp, err := vclient.Do(vreq)
	if err != nil {
	}
	defer vresp.Body.Close()
	bb, err := ioutil.ReadAll(vresp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var baseFields map[string]json.RawMessage
	var dataFields map[string]json.RawMessage
	err = json.Unmarshal(bb, &baseFields)
	if err != nil {
		fmt.Println(err)
	}
	var fields []VaultField
	for key, chunk := range baseFields {
		if key == "data" {
			err = json.Unmarshal(chunk, &dataFields)
			for fieldname, fieldvalue := range dataFields {
				var field VaultField
				field.Key = fieldname
				field.Value = strings.Replace(string(fieldvalue), "\"", "", -1)
				fields = append(fields, field)
			}
		}
	}
	var toreturn VaultSecret
	toreturn.Fields = fields
	return toreturn

}

func getToken() string {
	var vaulttoken string
	vaulttoken = getTokenViaEnv()
	if vaulttoken == "" {
		roleid := os.Getenv("VAULT_ROLEID")
		secretid := os.Getenv("VAULT_SECRETID")
		vaulttoken = getTokenViaApprole(roleid, secretid)
	}
	return vaulttoken
}
func getAddress() string {
	return os.Getenv("VAULT_ADDR")
}

func getTokenViaEnv() string {
	return os.Getenv("VAULT_TOKEN")
}

func getTokenViaApprole(roleid string, secretid string) string {
	return ""
}
