package ethcontracts

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArtifacts(t *testing.T) {
	workingPath, err := os.MkdirTemp("", "TestArtifacts")
	require.NoError(t, err)

	defer os.RemoveAll(workingPath)

	var (
		solFilePath      = "../../contractbinding/dummycontracts/TestContract.sol"
		newFilePath      = filepath.Join(workingPath, "artifact.json")
		originalArtifact *Artifact
		abiRaw           []byte
	)

	t.Run("CompileAndLoadContract", func(t *testing.T) {
		originalArtifact, abiRaw, err = CompileAndLoadContract(solFilePath, "")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(newFilePath, []byte(fmt.Sprintf(`
			{ "abi": %s, "bytecode": "%s" }
		`, abiRaw, hex.EncodeToString(originalArtifact.Bytecode))), 0770))
	})

	t.Run("LoadArtifacts", func(t *testing.T) {
		artifact, err := LoadArtifacts(workingPath, "artifact")

		require.NoError(t, err)
		require.Len(t, artifact, 1)
		require.Equal(t, originalArtifact.Abi, artifact["artifact"].Abi)
		require.Equal(t, originalArtifact.Bytecode, artifact["artifact"].Bytecode)
	})
}
