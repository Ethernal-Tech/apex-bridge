package ethcontracts

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// LoadArtifacts loads specified artifacts from desired directory
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

// CloneAndBuildContracts clones and builds smart contracts
// Note: git and npm must be in the path, repository must use hardhat
func CloneAndBuildContracts(
	dir, repositoryURL, repositoryName, artifactsDirName, branchName string,
) (string, error) {
	if _, err := executeCLICommand(
		"git", []string{"clone", "--progress", repositoryURL}, dir); err != nil {
		// git clone writes to stderror, check if messages are ok...
		// or if there is already existing git directory
		str := strings.TrimSpace(err.Error())
		if !strings.Contains(str, "Cloning into") && !strings.HasSuffix(str, "done.") &&
			!strings.Contains(str, fmt.Sprintf("'%s' already exists", repositoryName)) {
			return "", err
		}
	}

	dir = filepath.Join(dir, repositoryName)

	// do not listen for errors on following commands
	_, _ = executeCLICommand("git", []string{"checkout", branchName}, dir)
	_, _ = executeCLICommand("git", []string{"pull", "origin"}, dir)
	_, _ = executeCLICommand("npm", []string{"install"}, dir)

	if _, err := executeCLICommand("npx", []string{"hardhat", "compile"}, dir); err != nil {
		return "", err
	}

	return filepath.Join(dir, artifactsDirName), nil
}

// CompileContract compiles contract with solcjs and then loads artifact. DeployedBytecode will be nil
// Note: solcjs must be in the path
func CompileAndLoadContract(contractFilePath string, includePath string) (*Artifact, error) {
	workingPath, err := os.MkdirTemp("", "compile-and-load-contract")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(workingPath)

	fileName := filepath.Base(contractFilePath)
	directory := filepath.Dir(contractFilePath)
	fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	params := []string{
		"--optimize", "--bin", "--abi", "-o", workingPath, fileName,
	}
	if includePath != "" {
		params = append(params, "--include-path", includePath, "--base-path", ".")
	}

	if _, err := executeCLICommand("solcjs", params, directory); err != nil {
		return nil, err
	}

	parsedFileName := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(fileName, "_")
	binFilePath := filepath.Join(workingPath, fmt.Sprintf("%s_%s.bin", parsedFileName, fileNameWithoutExt))
	abiFilePath := filepath.Join(workingPath, fmt.Sprintf("%s_%s.abi", parsedFileName, fileNameWithoutExt))

	jsonRaw, err := os.ReadFile(abiFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load generated abi file: %s", abiFilePath)
	}

	hexBin, err := os.ReadFile(binFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load generated bin file: %s", binFilePath)
	}

	abi := new(abi.ABI)
	if err := json.Unmarshal(jsonRaw, &abi); err != nil {
		return nil, err
	}

	bytecode, err := hex.DecodeString(string(hexBin))
	if err != nil {
		return nil, err
	}

	return &Artifact{
		Abi:      abi,
		Bytecode: bytecode,
	}, nil
}

func executeCLICommand(binary string, args []string, workingDir string) (string, error) {
	var (
		stdErrBuffer bytes.Buffer
		stdOutBuffer bytes.Buffer
	)

	cmd := exec.Command(binary, args...)
	cmd.Stderr = &stdErrBuffer
	cmd.Stdout = &stdOutBuffer
	cmd.Dir = workingDir

	err := cmd.Run()

	if stdErrBuffer.Len() > 0 {
		return "", fmt.Errorf("error while executing command: %s", stdErrBuffer.String())
	} else if err != nil {
		return "", err
	}

	return stdOutBuffer.String(), nil
}
