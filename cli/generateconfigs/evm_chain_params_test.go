package cligenerateconfigs

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func newEvmExecuteParams(t *testing.T, args ...string) *evmChainGenerateConfigsParams {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "vc.json"), []byte(`{}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "relayer.json"), []byte(`{}`), 0o600))

	p := &evmChainGenerateConfigsParams{}
	cmd := &cobra.Command{}

	p.setFlags(cmd)
	require.NoError(t, cmd.ParseFlags(append([]string{
		"--" + chainIDStringFlag, "nexus",
		"--" + evmChainNodeURLFlag, "http://localhost:8545",
		"--" + relayerDataDirFlag, "relayer",
		"--" + outputDirFlag, dir,
		"--" + outputValidatorComponentsFileNameFlag, "vc.json",
		"--" + outputRelayerFileNameFlag, "relayer.json",
	}, args...)))

	return p
}

func readEvmChainConfig(t *testing.T, p *evmChainGenerateConfigsParams) map[string]any {
	t.Helper()

	vc := readJSONFile(t, filepath.Join(p.outputDir, p.outputValidatorComponentsFileName))

	return vc["ethChains"].(map[string]any)[p.chainIDString].(map[string]any)
}

func TestEvmChainParams_NoBatchPeriodPercent(t *testing.T) {
	t.Parallel()

	t.Run("default when flag is not set", func(t *testing.T) {
		t.Parallel()

		p := newEvmExecuteParams(t)

		require.NoError(t, p.validateFlags())

		_, err := p.Execute(nil)
		require.NoError(t, err)
		require.Equal(t, defaultEvmNoBatchPeriodPercent, readEvmChainConfig(t, p)["noBatchPeriodPercent"])
	})

	for _, value := range []string{"0", "0.1", "0.49"} {
		t.Run("valid "+value, func(t *testing.T) {
			t.Parallel()

			p := newEvmExecuteParams(t, "--"+evmNoBatchPeriodPercentFlag, value)

			require.NoError(t, p.validateFlags())

			_, err := p.Execute(nil)
			require.NoError(t, err)
			require.Equal(t, p.evmNoBatchPeriodPercent, readEvmChainConfig(t, p)["noBatchPeriodPercent"])
		})
	}

	for _, value := range []float64{-0.1, math.NaN()} {
		t.Run(fmt.Sprintf("invalid %v", value), func(t *testing.T) {
			t.Parallel()

			p := newEvmExecuteParams(t)
			p.evmNoBatchPeriodPercent = value

			require.ErrorContains(t, p.validateFlags(), evmNoBatchPeriodPercentFlag)
		})
	}
}
