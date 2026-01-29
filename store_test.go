package main

import "testing"

func TestStore_NewGame_InitialState(t *testing.T) {
	s := NewGameStore([]string{"APPLE"})

	st, err := s.NewGame()
	if err != nil {
		t.Fatalf("NewGame error: %v", err)
	}

	if st.GuessesRemaining != 6 {
		t.Fatalf("expected 6, got %d", st.GuessesRemaining)
	}
	if st.Current != "_____" {
		t.Fatalf("expected _____, got %q", st.Current)
	}
	if st.ID == "" {
		t.Fatalf("expected non-empty id")
	}
}

func TestStore_MakeGuess_CorrectAndWrong(t *testing.T) {
	s := NewGameStore([]string{"APPLE"})

	st, err := s.NewGame()
	if err != nil {
		t.Fatalf("NewGame error: %v", err)
	}

	// correct
	st2, err := s.MakeGuess(st.ID, 'P')
	if err != nil {
		t.Fatalf("MakeGuess error: %v", err)
	}
	if st2.Current != "_PP__" {
		t.Fatalf("expected _PP__, got %q", st2.Current)
	}
	if st2.GuessesRemaining != 6 {
		t.Fatalf("expected 6, got %d", st2.GuessesRemaining)
	}

	// wrong
	st3, err := s.MakeGuess(st.ID, 'Z')
	if err != nil {
		t.Fatalf("MakeGuess error: %v", err)
	}
	if st3.Current != "_PP__" {
		t.Fatalf("expected _PP__, got %q", st3.Current)
	}
	if st3.GuessesRemaining != 5 {
		t.Fatalf("expected 5, got %d", st3.GuessesRemaining)
	}
}

func TestStore_DeletesGameAfterCompletion(t *testing.T) {
	s := NewGameStore([]string{"A"})

	st, err := s.NewGame()
	if err != nil {
		t.Fatalf("NewGame error: %v", err)
	}

	// win in one guess
	_, err = s.MakeGuess(st.ID, 'A')
	if err != nil {
		t.Fatalf("MakeGuess error: %v", err)
	}

	// now should be gone
	_, err = s.MakeGuess(st.ID, 'A')
	if err != ErrGameNotFound {
		t.Fatalf("expected ErrGameNotFound, got %v", err)
	}
}
