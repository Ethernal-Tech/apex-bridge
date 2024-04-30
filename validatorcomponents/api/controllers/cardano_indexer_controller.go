package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

type CardanoIndexerControllerImpl struct {
	indexerDbs map[string]indexer.Database
	logger     hclog.Logger
}

var _ core.APIController = (*CardanoIndexerControllerImpl)(nil)

func NewCardanoIndexerController(
	indexerDbs map[string]indexer.Database, logger hclog.Logger,
) *CardanoIndexerControllerImpl {
	return &CardanoIndexerControllerImpl{
		indexerDbs: indexerDbs,
		logger:     logger,
	}
}

func (*CardanoIndexerControllerImpl) GetPathPrefix() string {
	return "CardanoIndexer"
}

func (c *CardanoIndexerControllerImpl) GetEndpoints() []*core.APIEndpoint {
	return []*core.APIEndpoint{
		{Path: "GetConfirmedTx", Method: http.MethodGet, Handler: c.getConfirmedTx, APIKeyAuth: true},
	}
}

func (c *CardanoIndexerControllerImpl) getConfirmedTx(w http.ResponseWriter, r *http.Request) {
	c.logger.Debug("getConfirmedTx called", "url", r.URL)

	queryValues := r.URL.Query()
	c.logger.Debug("query values", queryValues, "url", r.URL)

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		c.logger.Debug("chainId missing from query", "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, "chainId missing from query")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	txHashArr, exists := queryValues["txHash"]
	if !exists || len(txHashArr) == 0 {
		c.logger.Debug("txHash missing from query", "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, "txHash missing from query")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	chainID := chainIDArr[0]
	txHash := txHashArr[0]

	indexerDB, exists := c.indexerDbs[chainID]
	if !exists {
		c.logger.Debug("no indexer db for chainId", "chainId", chainID, "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, "no indexer db for chainId")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	tx, err := indexerDB.GetUnprocessedTx(txHash)
	if err != nil {
		c.logger.Debug(err.Error(), "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	if tx == nil {
		tx, err = indexerDB.GetProcessedTx(txHash)
		if err != nil {
			c.logger.Debug(err.Error(), "url", r.URL)

			rerr := utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			if rerr != nil {
				c.logger.Error("error while WriteErrorResponse", "err", rerr)
			}

			return
		}
	}

	if tx == nil {
		c.logger.Debug("Not found", "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusNotFound, "Not found")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	c.logger.Debug("getConfirmedTx success", "url", r.URL)

	err = json.NewEncoder(w).Encode(response.NewCardanoIndexerTxResponse(tx))
	if err != nil {
		c.logger.Error("error while writing response", "err", err)
	}
}
