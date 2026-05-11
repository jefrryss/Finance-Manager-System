package domain

import "testing"

func TestNewUserSuccess(t *testing.T) {
	user, err := NewUser("test@example.com", "tester", "hash")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user.Email != "test@example.com" || user.Login != "tester" {
		t.Fatalf("unexpected user payload")
	}
}

func TestNewUserInvalidEmail(t *testing.T) {
	_, err := NewUser("bad", "tester", "hash")
	if err != ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

