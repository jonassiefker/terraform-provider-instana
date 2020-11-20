package restapi

import (
	"encoding/json"
	"fmt"
)

func NewSyntheticEndpointsUnmarshaller() Unmarshaller {
	return &syntheticEndpointsUnmarshaller{}
}

type syntheticEndpointsUnmarshaller struct{}

func (u *syntheticEndpointsUnmarshaller) Unmarshal(data []byte) (InstanaDataObject, error) {
	syntheticEndpoints := SyntheticEndpoints{}
	if err := json.Unmarshal(data, &syntheticEndpoints); err != nil {
		return syntheticEndpoints, fmt.Errorf("failed to parse json; %s", err)
	}
	return syntheticEndpoints, nil
}
