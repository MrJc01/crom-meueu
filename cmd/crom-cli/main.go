package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"meueu/pkg/sdk"
)

func main() {
	// Subcommands
	loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
	postCmd := flag.NewFlagSet("post", flag.ExitOnError)
	dmCmd := flag.NewFlagSet("dm", flag.ExitOnError)
	// inboxCmd := flag.NewFlagSet("inbox", flag.ExitOnError)

	// Flags
	serverPtr := flag.String("server", "http://localhost:8080", "Crom Node URL")
	seedPtr := flag.String("seed", "", "Your Seed Phrase (Brain Key)")

	if len(os.Args) < 2 {
		fmt.Println("Usage: crom-cli <command> [flags]")
		fmt.Println("Commands: login, post, dm, inbox")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "login":
		loginCmd.Parse(os.Args[2:])
		if *seedPtr == "" {
			fmt.Println("Please provide -seed")
			return
		}
		auth := sdk.NewCromAuth()
		pub, err := auth.Login(*seedPtr)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("🔑 Logged in!\nPublic Key: %s\n", pub)

	case "post":
		postCmd.Parse(os.Args[2:])
		if *seedPtr == "" {
			log.Fatal("Please provide -seed")
		}
		content := postCmd.Arg(0)
		if content == "" {
			log.Fatal("Please provide content text")
		}

		auth := sdk.NewCromAuth()
		pub, _ := auth.Login(*seedPtr)
		client := sdk.NewClient(*serverPtr)

		// Create Payload
		payload := map[string]interface{}{
			"content":   content,
			"timestamp": time.Now().UnixMilli(),
		}
		payloadBytes, _ := json.Marshal(payload)

		// Canonical String (Simplified for CLI, should use strict canonicalizer in prod)
		// SDK `Sign` expects the canonical string.
		// "version:1|author:PUB|kind:text|timestamp:TS|payload:JSON"
		ts := time.Now().Unix() // Seconds
		kind := "text"
		canonical := fmt.Sprintf("version:1|author:%s|kind:%s|timestamp:%d|payload:%s", pub, kind, ts, string(payloadBytes))

		sig := auth.Sign(canonical)

		node := &sdk.Node{
			AuthorPubkey: pub,
			Kind:         kind,
			Payload:      json.RawMessage(payloadBytes),
			Signature:    sig,
			ClaimedAt:    time.Now().Format(time.RFC3339),
		}

		if err := client.Publish(node); err != nil {
			log.Fatal(err)
		}
		fmt.Println("✅ Posted successfully!")

	case "dm":
		dmCmd.Parse(os.Args[2:])
		to := dmCmd.Arg(0)
		msg := dmCmd.Arg(1)

		if to == "" || msg == "" {
			log.Fatal("Usage: crom-cli dm -seed=... <recipient_pub> <message>")
		}

		auth := sdk.NewCromAuth()
		pub, _ := auth.Login(*seedPtr)
		client := sdk.NewClient(*serverPtr)

		// Encrypt
		cipher, nonce, err := auth.EncryptDM(to, msg)
		if err != nil {
			log.Fatal("Encryption failed:", err)
		}

		payload := map[string]string{
			"ciphertext":       cipher,
			"nonce":            nonce,
			"recipient_pubkey": to,
		}
		payloadBytes, _ := json.Marshal(payload)

		ts := time.Now().Unix()
		kind := "encrypted_dm"
		canonical := fmt.Sprintf("version:1|author:%s|kind:%s|timestamp:%d|payload:%s", pub, kind, ts, string(payloadBytes))
		sig := auth.Sign(canonical)

		node := &sdk.Node{
			AuthorPubkey: pub,
			Kind:         kind,
			Payload:      json.RawMessage(payloadBytes),
			Tags:         []string{"recipient:" + to},
			Signature:    sig,
			ClaimedAt:    time.Now().Format(time.RFC3339),
		}

		if err := client.Publish(node); err != nil {
			log.Fatal(err)
		}
		fmt.Println("🔒 DM Sent!")

	default:
		fmt.Println("Unknown command")
	}
}
