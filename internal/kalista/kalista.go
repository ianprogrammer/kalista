package kalista

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/santhosh-tekuri/jsonschema"
	"gopkg.in/yaml.v2"
)

type contractDefinition struct {
	ContractId     string `yaml:"contractId"`
	Url            string
	Method         string
	Status         int
	Payload        string
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

	if len(contract.Request) > 0 {
		requestURL := contractId + "request.json"
		if err = compiler.AddResource(requestURL, strings.NewReader(contract.Request)); err != nil {
			return contractDefinition{}, err
		}
		requestSchema, err := compiler.Compile(requestURL)
		if err != nil {
			return contractDefinition{}, err
		}
		contract.requestSchema = requestSchema
	}
	if len(contract.Response) > 0 {
		responseURL := contractId + "response.json"
		if err = compiler.AddResource(responseURL, strings.NewReader(contract.Response)); err != nil {
			return contractDefinition{}, err
		}
		responseSchema, err := compiler.Compile(responseURL)
		if err != nil {
			return contractDefinition{}, err
		}
		contract.responseSchema = responseSchema
	}

	return contract, nil
}

type Kalista struct {
	yamls map[string][]byte
}

const succeed = "\u2713"
const failed = "\u2717"

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
					bytesFile, err := ioutil.ReadFile(path)
					if err != nil {
						log.Println(err)
					}
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

func (k *Kalista) StartContracTest() {
	var wg sync.WaitGroup

	wg.Add(len(k.yamls))
	for key := range k.yamls {
		go func(key string) {
			contract, err := k.buildContractDefinition(key)
			defer wg.Done()
			if err != nil {
				log.Println(err)
			}
			ok, err := k.MakeRequest(contract)
			printMessage(contract, err, ok, key, true)

		}(key)
	}
	wg.Wait()

}

func printMessage(contract contractDefinition, err error, ok bool, key string, isVerbose bool) {
	if err != nil {
		log.Printf("\t%s\t ContractId: %s, ContractPath: %s", failed, contract.ContractId, key)
		if isVerbose {
			log.Println(err)
		}
	}
	if ok {
		log.Printf("\t%s\t ContractId: %s, ContractPath: %s", succeed, contract.ContractId, key)
	}
}

func (k *Kalista) MakeRequest(c contractDefinition) (bool, error) {

	req, err := http.NewRequest(c.Method, c.Url, nil)

	if err != nil {
		return false, err
	}

	if c.requestSchema != nil && c.Payload != "" {
		req, err = http.NewRequest(c.Method, c.Url, bytes.NewBuffer([]byte(c.Payload)))
	}

	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode != c.Status {
		return false, err
	}

	var result interface{}
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return false, err
	}

	if c.requestSchema != nil && c.Payload != "" {

		var payload interface{}
		// validating response
		err = json.Unmarshal([]byte(c.Payload), &payload)

		if err != nil {
			return false, err
		}
		if err = c.requestSchema.ValidateInterface(payload); err != nil {
			return false, err
		}
	}

	// validating response
	err = json.Unmarshal(body, &result)

	if err != nil {
		return false, err
	}

	if err = c.responseSchema.ValidateInterface(result); err != nil {
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
