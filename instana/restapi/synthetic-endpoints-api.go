package restapi

import (
	"errors"

	"github.com/gessnerfl/terraform-provider-instana/utils"
)

const SyntheticEndpointsResourcePath = SettingsBasePath + "/synthetic-calls"

type MatchSpecification struct {
	Type     string `json:"type"`
	Key      string `json:"key"`
	Entity   string `json:"entity"`
	Value    string `json:"value"`
	Operator string `json:"operator"`
}

func (c MatchSpecification) Validate() error {
	if utils.IsBlank(c.Key) {
		return errors.New("Key is missing")
	}
	if utils.IsBlank(c.Entity) {
		return errors.New("Entity is missing")
	}
	if utils.IsBlank(c.Operator) {
		return errors.New("Operator is missing")
	}
	return nil
}

type CustomRule struct {
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	MatchSpecification MatchSpecification `json:"matchSpecification"`
	Enabled            bool               `json:"enabled"`
}

func (c CustomRule) Validate() error {
	if utils.IsBlank(c.Name) {
		return errors.New("Name is missing")
	}
	if len(c.Description) > 2048 {
		return errors.New("Description too long; Maximum number of characters in description is 2048")
	}
	return c.MatchSpecification.Validate()
}

type SyntheticEndpoints struct {
	ID                  string       `json:"id"`
	DefaultRulesEnabled bool         `json:"defaultRulesEnabled"`
	CustomRules         []CustomRule `json:"customRules"`
}

//GetID implemention of the interface InstanaDataObject
func (c SyntheticEndpoints) GetID() string {
	return c.ID
}

//Validate implementation of the interface InstanaDataObject to verify if data object is correct
func (c SyntheticEndpoints) Validate() error {
	if utils.IsBlank(c.ID) {
		return errors.New("ID is missing")
	}
	if len(c.CustomRules) > 500 {
		return errors.New("Too many CustomRules; Maximum number of CustomRules is 500")
	}
	for _, cr := range c.CustomRules {
		if err := cr.Validate(); err != nil {
			return err
		}
	}
	return nil
}
