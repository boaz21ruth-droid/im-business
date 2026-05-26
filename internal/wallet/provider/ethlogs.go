package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// erc20TransferTopic is the keccak256 of Transfer(address,address,uint256).
const erc20TransferTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

type ethLogEntry struct {
	Address         string   `json:"address"`
	Topics          []string `json:"topics"`
	Data            string   `json:"data"`
	BlockNumber     string   `json:"blockNumber"`
	TransactionHash string   `json:"transactionHash"`
}

// jsonRPCCall executes a standard JSON-RPC call and unmarshals the result.
func jsonRPCCall(ctx context.Context, client *http.Client, rpcURL, method string, params any, result any) error {
	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", rpcURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var wrapper struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return err
	}
	if wrapper.Error != nil {
		return fmt.Errorf("rpc %d: %s", wrapper.Error.Code, wrapper.Error.Message)
	}
	return json.Unmarshal(wrapper.Result, result)
}

// fetchERC20ByLogs fetches ERC20 transfer events for an address via eth_getLogs.
// Queries the last 50 000 blocks for both incoming and outgoing transfers.
// Native transfers are not supported via this method.
func fetchERC20ByLogs(ctx context.Context, client *http.Client, rpcURL, contractAddr, address string, limit int, source string) ([]TxRecord, error) {
	var latestHex string
	if err := jsonRPCCall(ctx, client, rpcURL, "eth_blockNumber", []any{}, &latestHex); err != nil {
		return nil, fmt.Errorf("eth_blockNumber: %w", err)
	}
	latest := hexToInt64(latestHex)
	fromBlock := latest - 50_000
	if fromBlock < 0 {
		fromBlock = 0
	}
	fromHex := fmt.Sprintf("0x%x", fromBlock)

	addrPadded := "0x000000000000000000000000" + strings.ToLower(strings.TrimPrefix(address, "0x"))

	var inLogs, outLogs []ethLogEntry
	inErr := jsonRPCCall(ctx, client, rpcURL, "eth_getLogs", []any{map[string]any{
		"fromBlock": fromHex, "toBlock": "latest",
		"address": contractAddr,
		"topics":  []any{erc20TransferTopic, nil, addrPadded},
	}}, &inLogs)
	outErr := jsonRPCCall(ctx, client, rpcURL, "eth_getLogs", []any{map[string]any{
		"fromBlock": fromHex, "toBlock": "latest",
		"address": contractAddr,
		"topics":  []any{erc20TransferTopic, addrPadded, nil},
	}}, &outLogs)

	if inErr != nil && outErr != nil {
		return nil, inErr
	}

	all := append(inLogs, outLogs...)
	seen := make(map[string]bool, len(all))
	records := make([]TxRecord, 0, len(all))
	for _, l := range all {
		if seen[l.TransactionHash] || len(l.Topics) < 3 {
			continue
		}
		seen[l.TransactionHash] = true
		from := topicToAddr(l.Topics[1])
		to := topicToAddr(l.Topics[2])
		contract := strings.ToLower(l.Address)
		records = append(records, TxRecord{
			Hash:          l.TransactionHash,
			From:          from,
			To:            to,
			Value:         hexToBigDecStr(l.Data),
			Decimals:      18,
			BlockNumber:   hexToInt64(l.BlockNumber),
			TokenContract: &contract,
			ChainKey:      "", // set by caller
			Source:        source,
		})
		if len(records) >= limit {
			break
		}
	}
	return records, nil
}

// topicToAddr extracts a 0x-prefixed address from a 32-byte padded topic.
func topicToAddr(topic string) string {
	if len(topic) >= 42 {
		return "0x" + topic[len(topic)-40:]
	}
	return topic
}
