package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"flow/internal/auth"
	"flow/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type mockUserStore struct {
	createUserCalls int
	createUserErr   error
	createdUser     *models.User
	getUser         *models.User
	getUserErr      error
	sessionCalls    int
	deleteCalls     int
}

func (m *mockUserStore) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
	m.createUserCalls++
	if m.createUserErr != nil {
		return nil, m.createUserErr
	}
	if m.createdUser != nil {
		return m.createdUser, nil
	}
	return &models.User{ID: "u1", Email: email, PasswordHash: passwordHash}, nil
}
func (m *mockUserStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.getUser, m.getUserErr
}
func (m *mockUserStore) CreateSession(ctx context.Context, sessionID, userID string, expiresAt time.Time) error {
	m.sessionCalls++
	return nil
}
func (m *mockUserStore) GetSession(ctx context.Context, sessionID string) (*models.Session, *models.User, error) {
	return nil, nil, auth.ErrSessionNotFound
}
func (m *mockUserStore) DeleteSession(ctx context.Context, sessionID string) error {
	m.deleteCalls++
	return nil
}

type mockSeeder struct {
	calls  int
	lastID string
}

func (m *mockSeeder) SeedDefaultCategories(ctx context.Context, userID string) error {
	m.calls++
	m.lastID = userID
	return nil
}

func TestRegister_PasswordTooShort(t *testing.T) {
	cases := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"7 chars rejected", "1234567", auth.ErrPasswordTooShort},
		{"empty rejected", "", auth.ErrPasswordTooShort},
		{"8 chars accepted", "12345678", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockUserStore{}
			seeder := &mockSeeder{}
			svc := auth.NewService(store, seeder, time.Hour)

			_, _, err := svc.Register(context.Background(), "user@example.com", tc.password)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				if store.createUserCalls != 0 {
					t.Fatalf("must not create user when password is invalid")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if seeder.calls != 1 || seeder.lastID != "u1" {
				t.Fatalf("seeder should be called once for new user, calls=%d id=%s", seeder.calls, seeder.lastID)
			}
		})
	}
}

func TestRegister_InvalidEmailRejected(t *testing.T) {
	store := &mockUserStore{}
	svc := auth.NewService(store, nil, time.Hour)

	_, _, err := svc.Register(context.Background(), "not-an-email", "password123")
	if !errors.Is(err, auth.ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
	if store.createUserCalls != 0 {
		t.Fatalf("must not create user for invalid email")
	}
}

func TestLogin_WrongPasswordRejected(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), 4)
	if err != nil {
		t.Fatal(err)
	}
	store := &mockUserStore{
		getUser: &models.User{ID: "u1", Email: "a@b.com", PasswordHash: string(hash)},
	}
	svc := auth.NewService(store, nil, time.Hour)

	_, _, err = svc.Login(context.Background(), "a@b.com", "wrong-password")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if store.sessionCalls != 0 {
		t.Fatalf("must not create session on bad password")
	}
}

func TestLogout_EmptySessionIsNoop(t *testing.T) {
	store := &mockUserStore{}
	svc := auth.NewService(store, nil, time.Hour)

	if err := svc.Logout(context.Background(), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.deleteCalls != 0 {
		t.Fatalf("empty session must not hit the store")
	}
}

func TestValidateSession_EmptyRejected(t *testing.T) {
	svc := auth.NewService(&mockUserStore{}, nil, time.Hour)
	_, err := svc.ValidateSession(context.Background(), "")
	if !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestLogin_SuccessCreatesSession(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), 4)
	if err != nil {
		t.Fatal(err)
	}
	store := &mockUserStore{
		getUser: &models.User{ID: "u1", Email: "a@b.com", PasswordHash: string(hash)},
	}
	svc := auth.NewService(store, nil, time.Hour)

	user, sessionID, err := svc.Login(context.Background(), "A@B.COM", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "u1" || sessionID == "" {
		t.Fatalf("expected user+session, got user=%v session=%q", user, sessionID)
	}
	if store.sessionCalls != 1 {
		t.Fatalf("expected one session create, got %d", store.sessionCalls)
	}
}
