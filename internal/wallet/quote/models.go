package quote

import "math/big"

// NativeTokenSentinel is the address aggregators use for a chain's native asset
// (ETH, BNB, …). Same convention across 0x, 1inch, OKX, Paraswap.
const NativeTokenSentinel = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"

// Token identifies an asset on a chain. ContractAddress == "" means native.
type Token struct {
	ContractAddress string // "" = native
}

// IsNative reports whether the token is the chain's native asset.
func (t Token) IsNative() bool { return t.ContractAddress == "" }

// APIAddress returns the address to send to an aggregator API.
func (t Token) APIAddress() string {
	if t.IsNative() {
		return NativeTokenSentinel
	}
	return t.ContractAddress
}

// Request is a normalized quote request, independent of any provider.
type Request struct {
	ChainKey    string
	SellToken   Token
	BuyToken    Token
	SellAmount  *big.Int
	Taker       string
	SlippageBps int
}

// Fees mirrors the per-quote fee breakdown the app's SwapFees model expects.
// All amounts are raw integer strings.
type Fees struct {
	IntegratorFeeAmount string `json:"integratorFeeAmount,omitempty"`
	IntegratorFeeToken  string `json:"integratorFeeToken,omitempty"`
	ZeroExFeeAmount     string `json:"zeroExFeeAmount,omitempty"`
	GasFeeAmount        string `json:"gasFeeAmount,omitempty"`
}

// Approval describes an ERC20 approval the client must perform before the swap.
type Approval struct {
	TokenAddress   string `json:"tokenAddress"`
	Spender        string `json:"spender"`
	RequiredAmount string `json:"requiredAmount"`
}

// PriceResult is a soft quote (preview only, no calldata).
type PriceResult struct {
	Provider    string `json:"provider"`
	BuyAmount   string `json:"buyAmount"`
	GasEstimate string `json:"gasEstimate,omitempty"`
	Fees        Fees   `json:"fees"`
}

// Quote is a firm quote with executable calldata. The client validates `To`
// against its router allow-list, signs locally, and broadcasts.
type Quote struct {
	Provider     string    `json:"provider"`
	BuyAmount    string    `json:"buyAmount"`
	MinBuyAmount string    `json:"minBuyAmount"`
	To           string    `json:"to"`
	Data         string    `json:"data"`
	Value        string    `json:"value"`
	Gas          string    `json:"gas,omitempty"`
	GasPrice     string    `json:"gasPrice,omitempty"`
	Approval     *Approval `json:"approval,omitempty"`
	Fees         Fees      `json:"fees"`
}

// PriceResponse is the aggregated soft-quote result: the winner plus the full
// ranked list (desc by buyAmount) for UI transparency.
type PriceResponse struct {
	Best *PriceResult   `json:"best"`
	All  []*PriceResult `json:"all"`
}

// QuoteResponse is the aggregated firm-quote result.
type QuoteResponse struct {
	Best *Quote   `json:"best"`
	All  []*Quote `json:"all"`
}
