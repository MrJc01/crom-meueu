package domain

type NodeFilter struct {
	IDs     []string `json:"ids,omitempty"`     // Filter by Specific IDs
	Kinds   []string `json:"kinds,omitempty"`   // Filter by Kind (e.g., "video", "text")
	Authors []string `json:"authors,omitempty"` // Filter by Author PubKey
	Tags    []string `json:"tags,omitempty"`    // Filter by Tags present in JSONB
}

type QueryRequest struct {
	Filters NodeFilter `json:"filters"`
	Limit   int        `json:"limit"`
	Offset  int        `json:"offset"`
}
