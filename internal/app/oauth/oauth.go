package oauthservice

import (
	"context"
	"errors"
	"fmt"

	"social-network/internal/domain/oauth"
	"social-network/internal/domain/user"
	"social-network/internal/pkg/uuid"

	oauthpkg "social-network/internal/pkg/oAuth"
)

type OAuthService struct {
	oauthRepo    oauth.Repository
	uuidProvider uuid.Provider
}

func NewOAuthService(oauthRepo oauth.Repository, uuidProvider uuid.Provider) *OAuthService {
	return &OAuthService{
		oauthRepo:    oauthRepo,
		uuidProvider: uuidProvider,
	}
}

func (s *OAuthService) Link(ctx context.Context, userID string, code string, provider oauthpkg.Provider) error {
	accessToken, err := provider.ExchangeCode(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	providerUserInfo, err := provider.GetUserInfo(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	providerID := providerUserInfo.ProviderID

	providerName := oauth.Provider(provider.Name())

	existingUser, err := s.oauthRepo.GetUserByProviderID(ctx, providerName, providerID)
	if err != nil && !errors.Is(err, oauth.ErrUserNotFound) {
		return fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil && existingUser.ID != userID {
		return oauth.ErrProviderAccountBelongsToAnotherUser
	}
	if existingUser != nil {
		return oauth.ErrAlreadyLinkedToProvider
	}
	// the check for userID,provider here is optional since we have the unique(userID,provider) in the DB table of oauthProvider

	oauthUser := &oauth.User{
		ProviderID: providerID,
		Provider:   providerName,
		Email:      providerUserInfo.Email,
		Username:   providerUserInfo.Username,
		AvatarURL:  providerUserInfo.AvatarURL,
		Name:       providerUserInfo.Name,
	}
	return s.oauthRepo.LinkOAuthProvider(ctx, userID, oauthUser)
}

func (s *OAuthService) Login(ctx context.Context, code string, provider oauthpkg.Provider) (*user.User, error) {
	accessToken, err := provider.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	providerUserInfo, err := provider.GetUserInfo(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	providerID := providerUserInfo.ProviderID

	// TODO: PROVIDER NAME VALIDATION
	providerName := oauth.Provider(provider.Name())
	existingUser, err := s.oauthRepo.GetUserByProviderID(ctx, providerName, providerID)
	if err != nil && !errors.Is(err, oauth.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return existingUser, nil
	}

	// //TO DO
	existingUser, err = s.oauthRepo.GetUserByEmail(ctx, providerUserInfo.Email)
	if err == nil && existingUser != nil {
		return existingUser, oauth.ErrUserWithEmailExists
	}

	oauthUser := &oauth.User{
		UserID:     s.uuidProvider.NewUUID(),
		ProviderID: providerID,
		Provider:   providerName,
		Email:      providerUserInfo.Email,
		Username:   providerUserInfo.Username,
		AvatarURL:  providerUserInfo.AvatarURL,
		Name:       providerUserInfo.Name,
	}

	newUser, err := s.oauthRepo.CreateOAuthUser(ctx, oauthUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, nil
}
