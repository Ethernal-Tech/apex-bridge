package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

type OracleStateControllerImpl struct {
	databases   map[string]indexer.Database
	adressesMap map[string][]string
	logger      hclog.Logger
}

var _ core.APIController = (*OracleStateControllerImpl)(nil)

func NewOracleStateController(
	databases map[string]indexer.Database,
	adressesMap map[string][]string,
	logger hclog.Logger,
) *OracleStateControllerImpl {
	return &OracleStateControllerImpl{
		databases:   databases,
		adressesMap: adressesMap,
		logger:      logger,
	}
}

func (*OracleStateControllerImpl) GetPathPrefix() string {
	return "OracleState"
}

func (c *OracleStateControllerImpl) GetEndpoints() []*core.APIEndpoint {
	return []*core.APIEndpoint{
		{Path: "Get", Method: http.MethodGet, Handler: c.getState, APIKeyAuth: true},
	}
}

func (c *OracleStateControllerImpl) getState(w http.ResponseWriter, r *http.Request) {
	c.logger.Debug("getState called", "url", r.URL)

	queryValues := r.URL.Query()

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		c.setError(w, r, "chainId missing from query")

		return
	}

	chainID := chainIDArr[0]

	db, existsDB := c.databases[chainID]
	addresses, existsAddrs := c.adressesMap[chainID]

	if !existsDB || !existsAddrs {
		c.setError(w, r, fmt.Sprintf("invalid chainID: %s", chainID))

		return
	}

	c.logger.Debug("getState success", "url", r.URL)

	w.Header().Set("Content-Type", "application/json")

	latestBlockPoint, err := db.GetLatestBlockPoint()
	if err != nil {
		c.setError(w, r, fmt.Sprintf("get latest point: %v", err))

		return
	}

	addressesUtxos := make([][]*indexer.TxInputOutput, len(addresses))
	count := 0

	for i, addr := range addresses {
		utxos, err := db.GetAllTxOutputs(addr, true)
		if err != nil {
			c.setError(w, r, fmt.Sprintf("get all tx outputs: %v", err))

			return
		}

		addressesUtxos[i] = utxos
		count += len(utxos)
	}

	outputUtxos := make([]oCore.CardanoChainConfigUtxo, 0, count)

	for _, utxos := range addressesUtxos {
		for _, inp := range utxos {
			outputUtxos = append(outputUtxos, oCore.CardanoChainConfigUtxo{
				Hash:    inp.Input.Hash,
				Index:   inp.Input.Index,
				Address: inp.Output.Address,
				Amount:  inp.Output.Amount,
				Slot:    inp.Output.Slot,
			})
		}
	}

	err = json.NewEncoder(w).Encode(response.NewOracleStateResponse(
		chainID, outputUtxos, latestBlockPoint.BlockSlot, latestBlockPoint.BlockHash))
	if err != nil {
		c.logger.Error("error while writing response", "err", err)
	}
}

func (c *OracleStateControllerImpl) setError(w http.ResponseWriter, r *http.Request, errString string) {
	c.logger.Debug("getState request", "err", errString, "url", r.URL)

	err := utils.WriteErrorResponse(w, http.StatusBadRequest, errString)
	if err != nil {
		c.logger.Error("error while WriteErrorResponse", "err", err)
	}
}
