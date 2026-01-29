package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T, words []string) *httptest.Server {
	t.Helper()

	app := &App{store: NewGameStore(words)}

	mux := http.NewServeMux()
	mux.HandleFunc("/new", app.handleNew)
	mux.HandleFunc("/guess", app.handleGuess)

	return httptest.NewServer(mux)
}

func postJSON(t *testing.T, client *http.Client, url string, body any) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode json: %v", err)
		}
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func TestNewGame_ReturnsInitialState(t *testing.T) {
	ts := newTestServer(t, []string{"APPLE"})
	defer ts.Close()

	resp := postJSON(t, ts.Client(), ts.URL+"/new", map[string]any{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got NewGameResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if strings.TrimSpace(got.ID) == "" {
		t.Fatalf("expected non-empty id")
	}
	if got.GuessesRemaining != 6 {
		t.Fatalf("expected guesses_remaining=6, got %d", got.GuessesRemaining)
	}
	if got.Current != "_____" {
		t.Fatalf("expected current=_____, got %q", got.Current)
	}
}

func TestGuess_CorrectLetter_DoesNotDecrement(t *testing.T) {
	ts := newTestServer(t, []string{"APPLE"})
	defer ts.Close()

	// new game
	respNew := postJSON(t, ts.Client(), ts.URL+"/new", map[string]any{})
	defer respNew.Body.Close()

	var ng NewGameResponse
	if err := json.NewDecoder(respNew.Body).Decode(&ng); err != nil {
		t.Fatalf("decode new: %v", err)
	}

	// guess P
	respGuess := postJSON(t, ts.Client(), ts.URL+"/guess", GuessRequest{ID: ng.ID, Guess: "P"})
	defer respGuess.Body.Close()

	if respGuess.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", respGuess.StatusCode)
	}

	var gr GuessResponse
	if err := json.NewDecoder(respGuess.Body).Decode(&gr); err != nil {
		t.Fatalf("decode guess: %v", err)
	}

	if gr.Current != "_PP__" {
		t.Fatalf("expected _PP__, got %q", gr.Current)
	}
	if gr.GuessesRemaining != 6 {
		t.Fatalf("expected guesses_remaining=6, got %d", gr.GuessesRemaining)
	}
}

func TestGuess_WrongLetter_Decrements(t *testing.T) {
	ts := newTestServer(t, []string{"APPLE"})
	defer ts.Close()

	respNew := postJSON(t, ts.Client(), ts.URL+"/new", map[string]any{})
	defer respNew.Body.Close()

	var ng NewGameResponse
	if err := json.NewDecoder(respNew.Body).Decode(&ng); err != nil {
		t.Fatalf("decode new: %v", err)
	}

	respGuess := postJSON(t, ts.Client(), ts.URL+"/guess", GuessRequest{ID: ng.ID, Guess: "Z"})
	defer respGuess.Body.Close()

	if respGuess.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", respGuess.StatusCode)
	}

	var gr GuessResponse
	if err := json.NewDecoder(respGuess.Body).Decode(&gr); err != nil {
		t.Fatalf("decode guess: %v", err)
	}

	if gr.Current != "_____" {
		t.Fatalf("expected _____, got %q", gr.Current)
	}
	if gr.GuessesRemaining != 5 {
		t.Fatalf("expected guesses_remaining=5, got %d", gr.GuessesRemaining)
	}
}

func TestGuess_InvalidGuess_Returns400(t *testing.T) {
	ts := newTestServer(t, []string{"APPLE"})
	defer ts.Close()

	respNew := postJSON(t, ts.Client(), ts.URL+"/new", map[string]any{})
	defer respNew.Body.Close()

	var ng NewGameResponse
	if err := json.NewDecoder(respNew.Body).Decode(&ng); err != nil {
		t.Fatalf("decode new: %v", err)
	}

	// invalid guess: 2 chars
	respGuess := postJSON(t, ts.Client(), ts.URL+"/guess", GuessRequest{ID: ng.ID, Guess: "AB"})
	defer respGuess.Body.Close()

	if respGuess.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", respGuess.StatusCode)
	}
}

func TestGuess_GameNotFound_Returns404(t *testing.T) {
	ts := newTestServer(t, []string{"APPLE"})
	defer ts.Close()

	respGuess := postJSON(t, ts.Client(), ts.URL+"/guess", GuessRequest{ID: "not-a-real-id", Guess: "A"})
	defer respGuess.Body.Close()

	if respGuess.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", respGuess.StatusCode)
	}
}

func TestGuess_GameIsDeletedAfterWin(t *testing.T) {
	ts := newTestServer(t, []string{"APPLE"})
	defer ts.Close()

	// new game
	respNew := postJSON(t, ts.Client(), ts.URL+"/new", map[string]any{})
	defer respNew.Body.Close()

	var ng NewGameResponse
	if err := json.NewDecoder(respNew.Body).Decode(&ng); err != nil {
		t.Fatalf("decode new: %v", err)
	}

	// Win with guesses: A, P, L, E
	letters := []string{"A", "P", "L", "E"}
	var last GuessResponse

	for _, l := range letters {
		respGuess := postJSON(t, ts.Client(), ts.URL+"/guess", GuessRequest{ID: ng.ID, Guess: l})
		if respGuess.StatusCode != http.StatusOK {
			respGuess.Body.Close()
			t.Fatalf("expected 200 for guess %s, got %d", l, respGuess.StatusCode)
		}
		if err := json.NewDecoder(respGuess.Body).Decode(&last); err != nil {
			respGuess.Body.Close()
			t.Fatalf("decode guess: %v", err)
		}
		respGuess.Body.Close()
	}

	if last.Current != "APPLE" {
		t.Fatalf("expected APPLE, got %q", last.Current)
	}

	// Next guess should be 404 because game was cleared
	respAfter := postJSON(t, ts.Client(), ts.URL+"/guess", GuessRequest{ID: ng.ID, Guess: "Z"})
	defer respAfter.Body.Close()

	if respAfter.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after game completed, got %d", respAfter.StatusCode)
	}
}
