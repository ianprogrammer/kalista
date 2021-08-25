package main

import (
	"github.com/santhosh-tekuri/jsonschema/v4"
	"gopkg.in/yaml.v2"
	"strings"
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
	RepoUrl   string
	RepoToken string
	yamls     map[string][]byte
}

func NewKalista() Kalista {
	return Kalista{}
}

func (k *Kalista) MakeTest(contractId string) (bool, error) {
	contract, err := k.buildContractDefinition(contractId)
	if err != nil {
		return false, err
	}
	//do request
	var res interface{}
	if err = contract.responseSchema.ValidateInterface(res); err != nil {
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
