package response

type HasTxFailedResponse struct {
	// true if the transaction failed, false otherwise
	Failed bool `json:"failed"`
} // @name HasTxFailedResponse
