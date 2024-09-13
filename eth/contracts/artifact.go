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
		return nil, fmt.Errorf("artifact found but no correct format: %w", err)
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
	mp := make(map[string]bool, len(names))
	result := make(map[string]*Artifact, len(names))

	for _, x := range names {
		mp[x+".json"] = true
	}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Print the file or directory path
		if mp[info.Name()] {
			a, err := LoadArtifactFromFile(path)
			if err != nil {
				return err
			}

			result[strings.TrimSuffix(info.Name(), ".json")] = a
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(result) != len(names) {
		return nil, fmt.Errorf("some artifacts not found: %d vs %d", len(result), len(names))
	}

	return result, nil
}
