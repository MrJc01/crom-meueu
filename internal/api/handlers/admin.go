package handlers

import (
	"encoding/json"
	"net/http"

	"meueu/internal/storage/postgres"
)

type AdminHandler struct {
	repo *postgres.AdminRepository
}

func NewAdminHandler(repo *postgres.AdminRepository) *AdminHandler {
	return &AdminHandler{repo: repo}
}

// GET /admin/stats
func (h *AdminHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.repo.GetStats(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// POST /admin/whitelist
// Input: { "public_key": "hex..." }
func (h *AdminHandler) HandleWhitelistAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.repo.AddToWhitelist(r.Context(), req.PublicKey); err != nil {
		http.Error(w, "Failed to add to whitelist", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /admin/whitelist
// Input: Query Param ?public_key=...
func (h *AdminHandler) HandleWhitelistRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pubKey := r.URL.Query().Get("public_key")
	if pubKey == "" {
		http.Error(w, "Missing public_key param", http.StatusBadRequest)
		return
	}

	if err := h.repo.RemoveFromWhitelist(r.Context(), pubKey); err != nil {
		http.Error(w, "Failed to remove from whitelist", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET /admin/whitelist
func (h *AdminHandler) HandleWhitelistList(w http.ResponseWriter, r *http.Request) {
	keys, err := h.repo.GetWhitelist(r.Context())
	if err != nil {
		http.Error(w, "Failed to list whitelist", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

// POST /admin/ban (Banned Words)
// Input: { "word": "badword" }
func (h *AdminHandler) HandleBanWord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Word string `json:"word"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.repo.AddBannedWord(r.Context(), req.Word); err != nil {
		http.Error(w, "Failed to add banned word", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) HandleUnbanWord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	word := r.URL.Query().Get("word")
	if word == "" {
		http.Error(w, "Missing word param", http.StatusBadRequest)
		return
	}

	if err := h.repo.RemoveBannedWord(r.Context(), word); err != nil {
		http.Error(w, "Failed to remove banned word", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) HandleListBannedWords(w http.ResponseWriter, r *http.Request) {
	words, err := h.repo.GetBannedWords(r.Context())
	if err != nil {
		http.Error(w, "Failed to list words", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(words)
}
