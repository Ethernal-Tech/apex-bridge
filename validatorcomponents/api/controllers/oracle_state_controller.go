package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	utxosMap := make(map[string][]*indexer.TxInputOutput, len(addresses))

	latestBlockPoint, err := db.GetLatestBlockPoint()
	if err != nil {
		c.setError(w, r, fmt.Sprintf("get latest point: %v", err))

		return
	}

	for _, addr := range addresses {
		utxos, err := db.GetAllTxOutputs(addr, true)
		if err != nil {
			c.setError(w, r, fmt.Sprintf("get all tx outputs: %v", err))

			return
		}

		utxosMap[addr] = utxos
	}

	err = json.NewEncoder(w).Encode(response.NewOracleStateResponse(chainID, utxosMap, latestBlockPoint))
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
