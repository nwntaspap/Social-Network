package usercommands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/domain/user"
	"social-network/internal/pkg/helpers"

	testhelpers "social-network/internal/pkg/testing"
)

func TestUserRegisterHandler_Handle(t *testing.T) {
	t.Run("group: user registration", func(t *testing.T) {
		testCases := newUserRegisterTestCases()
		for _, tt := range testCases {
			t.Run(tt.name, runUserRegisterTest(tt))
		}
	})
}

type userRegisterTestCase struct {
	name       string
	request    UserRegisterRequest
	setupMocks func(*testhelpers.MockRepository, *testhelpers.MockUUIDProvider, *testhelpers.MockEncryptionProvider)
	wantErr    error
	wantUser   *user.User
}

func newUserRegisterTestCases() []userRegisterTestCase {
	return []userRegisterTestCase{
		{
			name: "successful registration",
			request: UserRegisterRequest{
				Nickname: "testuser",
				Password: "password123",
				Email:    "test@example.com",
			},
			setupMocks: func(repo *testhelpers.MockRepository, uuid *testhelpers.MockUUIDProvider, enc *testhelpers.MockEncryptionProvider) {
				uuid.NewUUIDFunc = func() string { return "test-uuid" }
				enc.GenerateFunc = func(pass string) (string, error) { return "hashed_password", nil }
				repo.UserRegisterFunc = func(ctx context.Context, u *user.User) error { return nil }
			},
			wantErr: nil,
			wantUser: &user.User{
				ID:       "test-uuid",
				Nickname: "testuser",
				Email:    "test@example.com",
				Password: "hashed_password",
			},
		},
		{
			name: "encryption fails",
			request: UserRegisterRequest{
				Nickname: "testuser",
				Password: "password123",
				Email:    "test@example.com",
			},
			setupMocks: func(repo *testhelpers.MockRepository, uuid *testhelpers.MockUUIDProvider, enc *testhelpers.MockEncryptionProvider) {
				uuid.NewUUIDFunc = func() string { return "test-uuid" }
				enc.GenerateFunc = func(pass string) (string, error) { return "", testhelpers.ErrTest }
				repo.UserRegisterFunc = func(ctx context.Context, u *user.User) error { return nil }
			},
			wantErr:  testhelpers.ErrTest,
			wantUser: nil,
		},
		{
			name: "repository fails",
			request: UserRegisterRequest{
				Nickname: "testuser",
				Password: "password123",
				Email:    "test@example.com",
			},
			setupMocks: func(repo *testhelpers.MockRepository, uuid *testhelpers.MockUUIDProvider, enc *testhelpers.MockEncryptionProvider) {
				uuid.NewUUIDFunc = func() string { return "test-uuid" }
				enc.GenerateFunc = func(pass string) (string, error) { return "hashed_password", nil }
				repo.UserRegisterFunc = func(ctx context.Context, u *user.User) error { return testhelpers.ErrTest }
			},
			wantErr:  testhelpers.ErrTest,
			wantUser: nil,
		},
		{
			name: "email validation fails",
			request: UserRegisterRequest{
				Nickname: "testuser",
				Password: "password123",
				Email:    "invalid-email",
			},
			setupMocks: func(repo *testhelpers.MockRepository, uuid *testhelpers.MockUUIDProvider, enc *testhelpers.MockEncryptionProvider) {
				uuid.NewUUIDFunc = func() string { return "test-uuid" }
				enc.GenerateFunc = func(pass string) (string, error) { return "hashed_password", nil }
				repo.UserRegisterFunc = func(ctx context.Context, u *user.User) error { return testhelpers.ErrTest }
			},
			wantErr:  helpers.ErrInvalidEmailFormat,
			wantUser: nil,
		},
	}
}

func runUserRegisterTest(tt userRegisterTestCase) func(*testing.T) {
	return func(t *testing.T) {
		repo := &testhelpers.MockRepository{}
		uuid := &testhelpers.MockUUIDProvider{}
		enc := &testhelpers.MockEncryptionProvider{}
		tt.setupMocks(repo, uuid, enc)

		handler := NewUserRegisterHandler(repo, uuid, enc)
		got, err := handler.Handle(context.Background(), tt.request)

		if !errors.Is(err, tt.wantErr) {
			t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			return
		}

		testhelpers.AssertUserMatch(t, got, tt.wantUser)
	}
}

func TestNewUserRegisterHandler(t *testing.T) {
	repo := &testhelpers.MockRepository{}
	uuid := &testhelpers.MockUUIDProvider{}
	enc := &testhelpers.MockEncryptionProvider{}

	got := NewUserRegisterHandler(repo, uuid, enc)
	if got == nil {
		t.Fatal("NewUserRegisterHandler() returned nil")
	}
}
