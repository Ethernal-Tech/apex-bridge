package ethcontracts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

// DecodeArtifact unmarshals provided raw json content into an Artifact instance
func DecodeArtifact(data []byte) (*Artifact, error) {
	var hexRes HexArtifact

	if err := json.Unmarshal(data, &hexRes); err != nil {
		return nil, fmt.Errorf("artifact found but with incorrect format: %w", err)
	}

	bytecode, err := common.DecodeHex(hexRes.Bytecode)
	if err != nil {
		return nil, err
	}

	deployedBytecode, err := common.DecodeHex(hexRes.DeployedBytecode)
	if err != nil {
		return nil, err
	}

	return &Artifact{
		Abi:              hexRes.Abi,
		Bytecode:         bytecode,
		DeployedBytecode: deployedBytecode,
	}, nil
}

// LoadArtifactFromFile reads SC artifact file content and decodes it into an Artifact instance
func LoadArtifactFromFile(fileName string) (*Artifact, error) {
	jsonRaw, err := os.ReadFile(filepath.Clean(fileName))
	if err != nil {
		return nil, fmt.Errorf("failed to load artifact from file '%s': %w", fileName, err)
	}

	return DecodeArtifact(jsonRaw)
}

type HexArtifact struct {
	Abi              *abi.ABI `json:"abi"`
	Bytecode         string   `json:"bytecode"`
	DeployedBytecode string   `json:"deployedBytecode"`
}

type Artifact struct {
	Abi              *abi.ABI
	Bytecode         []byte
	DeployedBytecode []byte
}

func LoadArtifacts(directory string, names ...string) (map[string]*Artifact, error) {
	result := make(map[string]*Artifact, len(names))
	count := 0

	for _, x := range names {
		result[x] = nil
	}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fileName := strings.TrimSuffix(info.Name(), ".json")
		if value, exists := result[fileName]; exists && value == nil {
			artifact, err := LoadArtifactFromFile(path)
			if err != nil {
				return err
			}

			result[fileName] = artifact
			count++
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if count != len(names) {
		return nil, fmt.Errorf("some artifacts were not found: %d vs %d", count, len(names))
	}

	return result, nil
}
