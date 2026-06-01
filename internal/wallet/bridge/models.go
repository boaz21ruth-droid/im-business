package bridge

import "math/big"

// NativeToken is LI.FI's representation of a chain's native asset (zero address).
// NOTE: this differs from the DEX aggregators' 0xEee… sentinel.
const NativeToken = "0x0000000000000000000000000000000000000000"

// Request is a normalized cross-chain quote request. FromToken/ToToken == ""
// or "native" denote the native asset.
type Request struct {
	FromChain   string
	ToChain     string
	FromToken   string
	ToToken     string
	FromAmount  *big.Int
	FromAddress string
	ToAddress   string // defaults to FromAddress
	SlippageBps int
}

// Approval describes the ERC20 approval the client must perform on the source
// chain before the bridge tx. Nil for native-token sources.
type Approval struct {
	TokenAddress   string `json:"tokenAddress"`
	Spender        string `json:"spender"`
	RequiredAmount string `json:"requiredAmount"`
}

// Quote is a cross-chain route: a source-chain signable tx plus delivery metadata.
type Quote struct {
	Tool                 string    `json:"tool"`        // chosen bridge, e.g. "across"
	FromChain            string    `json:"fromChain"`
	ToChain              string    `json:"toChain"`
	ToAmount             string    `json:"toAmount"`    // expected received on dest
	ToAmountMin          string    `json:"toAmountMin"` // min after slippage
	To                   string    `json:"to"`          // source-chain tx target
	Data                 string    `json:"data"`
	Value                string    `json:"value"` // decimal (normalized from hex)
	Gas                  string    `json:"gas,omitempty"`
	GasPrice             string    `json:"gasPrice,omitempty"`
	Approval             *Approval `json:"approval,omitempty"`
	ExecutionDurationSec int       `json:"executionDurationSec"`
}

// Status is the cross-chain delivery progress.
type Status struct {
	Status     string `json:"status"` // PENDING / DONE / FAILED / NOT_FOUND / INVALID
	Substatus  string `json:"substatus,omitempty"`
	Message    string `json:"message,omitempty"`
	DestTxHash string `json:"destTxHash,omitempty"`
	DestAmount string `json:"destAmount,omitempty"`
	Explorer   string `json:"explorer,omitempty"` // LI.FI explorer link
}
