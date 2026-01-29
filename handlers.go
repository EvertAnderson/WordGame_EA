package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type App struct {
	store *GameStore
}

type NewGameResponse struct {
	ID               string `json:"id"`
	Current          string `json:"current"`
	GuessesRemaining int    `json:"guesses_remaining"`
}

type GuessRequest struct {
	ID    string `json:"id"`
	Guess string `json:"guess"`
}

type GuessResponse struct {
	ID               string `json:"id"`
	Current          string `json:"current"`
	GuessesRemaining int    `json:"guesses_remaining"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func isValidGuess(s string) bool {
	// Must be exactly one ASCII character between A-Z
	if len(s) != 1 {
		return false
	}
	b := s[0]
	return b >= 'A' && b <= 'Z'
}

func (app *App) handleNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	state, err := app.store.NewGame()
	if err != nil {
		http.Error(w, "failed to create game", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, NewGameResponse{
		ID:               state.ID,
		Current:          state.Current,
		GuessesRemaining: state.GuessesRemaining,
	})
}

func (app *App) handleGuess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req GuessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.ID = strings.TrimSpace(req.ID)
	req.Guess = strings.TrimSpace(strings.ToUpper(req.Guess))

	if req.ID == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	if !isValidGuess(req.Guess) {
		http.Error(w, "guess must be a single ASCII character [A-Z]", http.StatusBadRequest)
		return
	}

	guessRune := rune(req.Guess[0])

	state, err := app.store.MakeGuess(req.ID, guessRune)
	if err != nil {
		if err == ErrGameNotFound {
			http.Error(w, "game not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to apply guess", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, GuessResponse{
		ID:               state.ID,
		Current:          state.Current,
		GuessesRemaining: state.GuessesRemaining,
	})
}
