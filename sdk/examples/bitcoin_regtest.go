package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)
type BitcoinRPC struct {
	URL      string
	User     string
	Password string
}

type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func (b *BitcoinRPC) Call(method string, params ...interface{}) (interface{}, error) {
	req := RPCRequest{
		JSONRPC: "1.0",
		ID:      "meshguard-test",
		Method:  method,
		Params:  params,
	}

	body, _ := json.Marshal(req)

	resp, err := httpPost(b.URL, b.User, b.Password, body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	if result["error"] != nil && result["error"] != nil {
		return nil, fmt.Errorf("rpc error: %v", result["error"])
	}

	return result["result"], nil
}

/*
-----------------------
HTTP helper
-----------------------
*/
func httpPost(url, user, pass string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func main() {

	fmt.Println("=== REGTEST BITCOIN INTEGRATION TEST ===")

	// Wallet-specific RPC endpoints
	sender := &BitcoinRPC{
		URL:      "http://127.0.0.1:18443/wallet/meshguard",
		User:     "john",
		Password: "john",
	}

	receiver := &BitcoinRPC{
		URL:      "http://127.0.0.1:18443/wallet/meshgaurdprime",
		User:     "john",
		Password: "john",
	}

	// Step 1: Get receiver address
	addrRaw, err := receiver.Call("getnewaddress")
	if err != nil {
		log.Fatal(err)
	}
	receiverAddr := addrRaw.(string)

	fmt.Println("Receiver address:", receiverAddr)

	// Step 2: Send BTC from sender → receiver
	fmt.Println("\nSending 1.0 BTC...")

	txid, err := sender.Call("sendtoaddress", receiverAddr, 1.0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("TXID:", txid)

	// Step 3: Mine block to confirm transaction
	fmt.Println("\nMining block...")

	mineAddrRaw, _ := sender.Call("getnewaddress")
	mineAddr := mineAddrRaw.(string)

	_, err = sender.Call("generatetoaddress", 1, mineAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Step 4: Wait for confirmation
	time.Sleep(2 * time.Second)

	// Step 5: Check balances
	senderBal, _ := sender.Call("getbalance")
	receiverBal, _ := receiver.Call("getbalance")

	fmt.Println("\n=== FINAL RESULTS ===")
	fmt.Println("Sender balance:", senderBal)
	fmt.Println("Receiver balance:", receiverBal)

	fmt.Println("\n✓ REGTEST TRANSACTION COMPLETE")
}