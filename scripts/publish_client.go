//go:build ignore

package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	msgPtr := flag.String("msg", "Hello from CLI Client", "Message content to publish")
	kindPtr := flag.String("kind", "text", "Kind of the node (text, video, etc)")
	apiPtr := flag.String("api", "http://localhost:8080/v1/publish", "API Endpoint")
	flag.Parse()

	// 1. Generate Ephemeral Identity
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}
	pubKeyHex := hex.EncodeToString(pubKey)
	log.Printf("🔑 Identity Generated: %s", pubKeyHex)

	// 2. Construct Payload
	timestamp := time.Now()
	payloadMap := map[string]string{"content": *msgPtr}
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		log.Fatalf("Failed to marshal payload: %v", err)
	}

	// 3. Sign (Canonical String Logic must match Server)
	// Format: version:1|author:<pubkey>|kind:<kind>|timestamp:<claimed_at_unix>|payload:<payload_json_string>
	canonicalMsg := fmt.Sprintf("version:1|author:%s|kind:%s|timestamp:%d|payload:%s",
		pubKeyHex,
		*kindPtr,
		timestamp.Unix(),
		string(payloadBytes),
	)

	signature := ed25519.Sign(privKey, []byte(canonicalMsg))
	sigHex := hex.EncodeToString(signature)

	// 4. Send Request
	requestBody := map[string]interface{}{
		"author_pubkey": pubKeyHex,
		"kind":          *kindPtr,
		"payload":       json.RawMessage(payloadBytes),
		"signature":     sigHex,
		"claimed_at":    timestamp,
	}

	jsonBody, _ := json.Marshal(requestBody)
	resp, err := http.Post(*apiPtr, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 201 {
		log.Printf("✅ Success! Node Created.\nResponse: %s", string(respBytes))
	} else {
		log.Printf("❌ Error (Status %d): %s", resp.StatusCode, string(respBytes))
	}
}
