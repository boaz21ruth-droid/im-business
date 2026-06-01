package intent

import "math/big"

// Request is a normalized intent-quote request.
type Request struct {
	ChainKey    string
	SellToken   string
	BuyToken    string
	SellAmount  *big.Int
	From        string // signer / receiver
	SlippageBps int
}

// Order is the CoW GPv2Order the client signs (EIP-712). Field order/types match
// the Order struct hashed by the client.
type Order struct {
	SellToken         string `json:"sellToken"`
	BuyToken          string `json:"buyToken"`
	Receiver          string `json:"receiver"`
	SellAmount        string `json:"sellAmount"`
	BuyAmount         string `json:"buyAmount"`
	ValidTo           int64  `json:"validTo"`
	AppData           string `json:"appData"` // 32-byte hash
	FeeAmount         string `json:"feeAmount"`
	Kind              string `json:"kind"`
	PartiallyFillable bool   `json:"partiallyFillable"`
	SellTokenBalance  string `json:"sellTokenBalance"`
	BuyTokenBalance   string `json:"buyTokenBalance"`
}

// QuoteResult is returned to the client: the order to sign + everything needed
// to build the EIP-712 digest and the pre-trade approval.
type QuoteResult struct {
	Order             Order  `json:"order"`
	ChainID           int    `json:"chainId"`
	VerifyingContract string `json:"verifyingContract"` // GPv2Settlement
	ApprovalSpender   string `json:"approvalSpender"`   // GPv2VaultRelayer
	QuoteID           int64  `json:"quoteId"`
	ExpectedBuyAmount string `json:"expectedBuyAmount"` // pre-slippage estimate
}

// SubmitRequest carries a client-signed order for submission.
type SubmitRequest struct {
	Order     Order  `json:"order"`
	Signature string `json:"signature"` // 0x… 65-byte r‖s‖v
	From      string `json:"from"`
	QuoteID   int64  `json:"quoteId"`
}

// OrderStatus is an order's settlement state.
type OrderStatus struct {
	Status            string `json:"status"` // open/fulfilled/cancelled/expired
	ExecutedBuyAmount string `json:"executedBuyAmount,omitempty"`
}
