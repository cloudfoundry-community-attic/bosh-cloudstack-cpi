package applyspec

import (
	"encoding/json"

	bosherr "bosh/errors"
	boshsys "bosh/system"
)

type concreteV1Service struct {
	specFilePath string
	fs           boshsys.FileSystem
}

func NewConcreteV1Service(fs boshsys.FileSystem, specFilePath string) (service concreteV1Service) {
	service.fs = fs
	service.specFilePath = specFilePath
	return
}

func (s concreteV1Service) Get() (V1ApplySpec, error) {
	var spec V1ApplySpec

	if !s.fs.FileExists(s.specFilePath) {
		return spec, nil
	}

	contents, err := s.fs.ReadFile(s.specFilePath)
	if err != nil {
		return spec, bosherr.WrapError(err, "Reading json spec file")
	}

	err = json.Unmarshal([]byte(contents), &spec)
	if err != nil {
		return spec, bosherr.WrapError(err, "Unmarshalling json spec file")
	}

	return spec, nil
}

func (s concreteV1Service) Set(spec V1ApplySpec) error {
	specBytes, err := json.Marshal(spec)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling apply spec")
	}

	err = s.fs.WriteFile(s.specFilePath, specBytes)
	if err != nil {
		return bosherr.WrapError(err, "Writing spec to disk")
	}

	return nil
}
