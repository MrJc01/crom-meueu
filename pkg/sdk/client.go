package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client handles network interaction with Crom Node
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient(url string) *Client {
	if url == "" {
		url = "http://localhost:8080"
	}
	return &Client{
		BaseURL: url,
		HTTP:    &http.Client{Timeout: 10 * time.Second},
	}
}

type Node struct {
	ID           string          `json:"id,omitempty"`
	AuthorPubkey string          `json:"author_pubkey"`
	Kind         string          `json:"kind"`
	Payload      json.RawMessage `json:"payload"`
	Tags         []string        `json:"tags,omitempty"`
	Signature    string          `json:"signature"`
	ClaimedAt    string          `json:"claimed_at"`
	OriginServer string          `json:"origin_server,omitempty"`
}

type Filter struct {
	IDs     []string `json:"ids,omitempty"`
	Kinds   []string `json:"kinds,omitempty"`
	Authors []string `json:"authors,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

type QueryRequest struct {
	Filters Filter `json:"filters"`
	Limit   int    `json:"limit"`
}

// Publish sends a signed node to the network
func (c *Client) Publish(node *Node) error {
	data, err := json.Marshal(node)
	if err != nil {
		return err
	}

	resp, err := c.HTTP.Post(c.BaseURL+"/v1/publish", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("publish failed: %s", resp.Status)
	}

	return nil
}

// Query fetches nodes
func (c *Client) Query(filter Filter, limit int) ([]Node, error) {
	reqBody := QueryRequest{
		Filters: filter,
		Limit:   limit,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTP.Post(c.BaseURL+"/v1/query", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("query failed: %s", resp.Status)
	}

	var nodes []Node
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}
