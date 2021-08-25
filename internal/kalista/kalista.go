package kalista

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema"
	"gopkg.in/yaml.v2"
)

type contractDefinition struct {
	Url            string
	Method         string
	Request        string
	Response       string
	requestSchema  *jsonschema.Schema
	responseSchema *jsonschema.Schema
}

func (k *Kalista) buildContractDefinition(contractId string) (contractDefinition, error) {
	var contract contractDefinition
	err := yaml.Unmarshal(k.yamls[contractId], &contract)
	if err != nil {
		return contractDefinition{}, err
	}
	compiler := jsonschema.NewCompiler()
	requestURL := contractId + "request.json"

	if err = compiler.AddResource(requestURL, strings.NewReader(contract.Request)); err != nil {
		return contractDefinition{}, err
	}
	responseURL := contractId + "response.json"
	if err = compiler.AddResource(responseURL, strings.NewReader(contract.Response)); err != nil {
		return contractDefinition{}, err
	}

	requestSchema, err := compiler.Compile(requestURL)
	if err != nil {
		return contractDefinition{}, err
	}
	responseSchema, err := compiler.Compile(requestURL)
	if err != nil {
		return contractDefinition{}, err
	}
	contract.requestSchema = requestSchema
	contract.responseSchema = responseSchema

	return contract, nil
}

type Kalista struct {
	yamls map[string][]byte
}

func readAllContractFiles(path string) map[string][]byte {

	yamls := make(map[string][]byte)
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ext := filepath.Ext(path)
				if ext != "yml" && ext != "yaml" {
					bytesFile, _ := ioutil.ReadFile(path)
					yamls[path] = bytesFile
				}
			}
			return nil
		})

	if err != nil {
		log.Println(err)
	}

	return yamls

}

func NewKalista(folderPath string) Kalista {
	return Kalista{
		yamls: readAllContractFiles(folderPath),
	}
}

func (k *Kalista) MakeTest(contractId string) (bool, error) {
	contract, err := k.buildContractDefinition(contractId)
	if err != nil {
		return false, err
	}

	req, _ := http.NewRequest(contract.Method, contract.Url, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()
	var result interface{}
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return false, err
	}

	err = json.Unmarshal(body, &result)

	if err != nil {
		return false, err
	}

	if err = contract.responseSchema.ValidateInterface(result); err != nil {
		return false, err
	}

	return true, nil
}

func (k *Kalista) MakeTestWithBody(contractId string, body interface{}) (bool, error) {
	contract, err := k.buildContractDefinition(contractId)
	if err != nil {
		return false, err
	}
	if err = contract.requestSchema.ValidateInterface(body); err != nil {
		return false, err
	}
	//do request
	var res interface{}
	if err = contract.responseSchema.ValidateInterface(res); err != nil {
		return false, err
	}
	return true, nil
}
