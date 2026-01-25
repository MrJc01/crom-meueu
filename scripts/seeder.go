//go:build ignore

package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Content Pools
var videos = []map[string]string{
	{"url": "https://storage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4", "title": "Big Buck Bunny", "content": "A classic open movie project.", "thumbnail": "https://placehold.co/600x400/222/FFF?text=Big+Buck+Bunny"},
	{"url": "https://storage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4", "title": "Elephants Dream", "content": "The world's first open movie.", "thumbnail": "https://placehold.co/600x400/333/FFF?text=Elephants+Dream"},
	{"url": "https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4", "title": "For Bigger Blazes", "content": "Chromecast sample video.", "thumbnail": "https://placehold.co/600x400/444/FFF?text=For+Bigger+Blazes"},
	{"url": "https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerEscapes.mp4", "title": "For Bigger Escapes", "content": "Action packed escape.", "thumbnail": "https://placehold.co/600x400/555/FFF?text=For+Bigger+Escapes"},
	{"url": "https://storage.googleapis.com/gtv-videos-bucket/sample/Sintel.mp4", "title": "Sintel", "content": "Third open movie by Blender Foundation.", "thumbnail": "https://placehold.co/600x400/666/FFF?text=Sintel"},
}

var texts = []string{
	"Just setting up my distributed social node! #decentralized",
	"Anyone else thinks protocols > platforms? #web3",
	"Coding a new frontend for Crom... it's surprisingly easy.",
	"The future is modular. The future is sovereignty.",
	"Why are we still using centralized silos in 2026?",
	"Hello world! This message is signed by my private key.",
	"Golang + Postgres is a beast combination for performance.",
	"Just watched a great documentary on the protocol.",
	"Thinking about building a CLI client in Rust.",
	"Remember: Not your keys, not your identity.",
}

var articles = []map[string]string{
	{"title": "Why API-First Matters", "content": "In a world of diverse devices, the API is the only constant..."},
	{"title": "The Death of the Monolith", "content": "Breaking down social networks into composable parts is the next step..."},
	{"title": "Cryptography 101 for Devs", "content": "Understanding Ed25519 signatures is crucial for modern apps..."},
	{"title": "How to run your own Crom Node", "content": "A step-by-step guide to digital independence..."},
}

func main() {
	apiURL := "http://localhost:8080/v1/publish"
	totalNodes := 30

	fmt.Printf("🌱 Starting Seeder... Targeting %d nodes.\n", totalNodes)

	// Single Identity for the "Power User" (to show same author multiple times)
	pubKey, privKey, _ := ed25519.GenerateKey(nil)
	pubKeyHex := hex.EncodeToString(pubKey)

	for i := 0; i < totalNodes; i++ {
		kind := randomKind()
		payload := generatePayload(kind)

		// Sometimes use a new identity (random users)
		pk, sk, _ := ed25519.GenerateKey(nil)
		authorPub := hex.EncodeToString(pk)
		authorPriv := sk

		// 30% chance to be the Power User
		if rand.Intn(100) < 30 {
			authorPub = pubKeyHex
			authorPriv = privKey
		}

		publish(apiURL, authorPub, authorPriv, kind, payload)
		time.Sleep(100 * time.Millisecond) // Don't hammer too hard
	}

	fmt.Println("✅ Seeding Complete!")
}

func randomKind() string {
	r := rand.Intn(100)
	if r < 40 {
		return "text"
	} // 40% Text
	if r < 70 {
		return "video"
	} // 30% Video
	return "article" // 30% Article
}

func generatePayload(kind string) map[string]interface{} {
	switch kind {
	case "video":
		v := videos[rand.Intn(len(videos))]
		return map[string]interface{}{
			"url":       v["url"],
			"title":     v["title"],
			"content":   v["content"],
			"thumbnail": v["thumbnail"],
		}
	case "article":
		a := articles[rand.Intn(len(articles))]
		return map[string]interface{}{
			"title":   a["title"],
			"content": a["content"] + fmt.Sprintf(" (Random ID: %d)", rand.Int()),
		}
	default: // text
		return map[string]interface{}{
			"content": texts[rand.Intn(len(texts))],
		}
	}
}

func publish(url, pubKeyHex string, privKey ed25519.PrivateKey, kind string, payload map[string]interface{}) {
	payloadBytes, _ := json.Marshal(payload)
	timestamp := time.Now()

	// Canonical String
	canonicalMsg := fmt.Sprintf("version:1|author:%s|kind:%s|timestamp:%d|payload:%s",
		pubKeyHex, kind, timestamp.Unix(), string(payloadBytes),
	)

	signature := ed25519.Sign(privKey, []byte(canonicalMsg))
	sigHex := hex.EncodeToString(signature)

	requestBody := map[string]interface{}{
		"author_pubkey": pubKeyHex,
		"kind":          kind,
		"payload":       json.RawMessage(payloadBytes),
		"signature":     sigHex,
		"claimed_at":    timestamp,
	}

	jsonBody, _ := json.Marshal(requestBody)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Failed to post: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		log.Printf("Error posting %s: Status %d", kind, resp.StatusCode)
	} else {
		fmt.Printf("Published [%s] Node\n", kind)
	}
}
