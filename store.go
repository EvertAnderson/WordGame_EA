package main

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
)

var (
	ErrGameNotFound = errors.New("game not found")
)

type Game struct {
	ID               string
	Word             string
	Current          []rune
	GuessesRemaining int
}

type GameState struct {
	ID               string
	Current          string
	GuessesRemaining int
	Completed        bool
}

type GameStore struct {
	mu    sync.Mutex
	games map[string]*Game
	words []string
}

func NewGameStore(words []string) *GameStore {
	return &GameStore{
		games: make(map[string]*Game),
		words: words,
	}
}

func (s *GameStore) NewGame() (GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, err := generateIdentifier()
	if err != nil {
		return GameState{}, err
	}

	word := strings.ToUpper(s.words[rand.Intn(len(s.words))])
	current := make([]rune, len([]rune(word)))
	for i := range current {
		current[i] = '_'
	}

	g := &Game{
		ID:               id,
		Word:             word,
		Current:          current,
		GuessesRemaining: 6,
	}

	s.games[g.ID] = g

	return GameState{
		ID:               g.ID,
		Current:          string(g.Current),
		GuessesRemaining: g.GuessesRemaining,
		Completed:        false,
	}, nil
}

// MakeGuess - Applies a guess to the game and returns an updated state.
// If the game is completed (win/lose), it is removed from the store.
func (s *GameStore) MakeGuess(id string, guess rune) (GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.games[id]
	if !ok {
		return GameState{}, ErrGameNotFound
	}

	found := false
	wordRunes := []rune(g.Word)

	for i := range wordRunes {
		if wordRunes[i] == guess {
			g.Current[i] = guess
			found = true
		}
	}

	if !found {
		g.GuessesRemaining--
		if g.GuessesRemaining < 0 {
			g.GuessesRemaining = 0
		}
	}

	// Determine completion
	completed := false
	if string(g.Current) == g.Word {
		completed = true // win
	}
	if g.GuessesRemaining == 0 {
		completed = true // lose
	}

	state := GameState{
		ID:               g.ID,
		Current:          string(g.Current),
		GuessesRemaining: g.GuessesRemaining,
		Completed:        completed,
	}

	// Completed games are cleared from the data store
	if completed {
		delete(s.games, g.ID)
	}

	return state, nil
}
