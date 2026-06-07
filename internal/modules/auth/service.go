package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github/sanjay-khandelwal/internal/modules/auth/dto"
	"github/sanjay-khandelwal/internal/modules/user"
	userDto "github/sanjay-khandelwal/internal/modules/user/dto"
	"github/sanjay-khandelwal/internal/shared/apperr"
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"
	"github/sanjay-khandelwal/internal/shared/security"
	"github/sanjay-khandelwal/internal/shared/security/jwt"
	"github/sanjay-khandelwal/internal/shared/security/password"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Service is the interface handlers depend on.
type Service interface {

	// registration process
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error)
	CreateUser(ctx context.Context, tx pgx.Tx, req *dto.RegisterRequest) (uuid.UUID, error)

	//email verification process
	SendVerification(ctx context.Context, tx pgx.Tx, userID uuid.UUID, email string) error
	EmailVerify(ctx context.Context, token string) error
	validateVerificationToken(verification *EmailVerification, token string) error

	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error)
}

type service struct {
	repo        Repository
	hasher      *password.Hasher
	jwt         *jwt.Service
	userservice user.Service
	posgressDb  *postgres.DB
}

func NewService(repo Repository, db *postgres.DB, hasher *password.Hasher, jwt *jwt.Service, userservice user.Service) Service {
	return &service{repo: repo, posgressDb: db, hasher: hasher, jwt: jwt, userservice: userservice}
}

// Register creates a new user + profile inside a single transaction.
//
// Flow:
//  1. Hash password (Argon2id)
//  2. Begin transaction
//  3. INSERT users — DB unique constraint returns 409 on duplicate email (no TOCTOU)
//  4. INSERT profiles
//  5. Commit → return response
func (s *service) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	var resp *dto.UserResponse

	err := s.posgressDb.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		userID, err := s.CreateUser(ctx, tx, req)
		if err != nil {
			return err
		}

		if err := s.userservice.CreateProfile(ctx, tx, &userDto.CreateProfileRequest{
			UserID:    userID.String(),
			FirstName: req.FirstName,
			LastName:  req.LastName,
		}); err != nil {
			return err
		}

		if err := s.SendVerification(ctx, tx, userID, req.Email); err != nil {
			return err
		}

		resp = &dto.UserResponse{
			ID:        userID.String(),
			Email:     req.Email,
			CreatedAt: time.Now().UTC(),
		}
		return nil
	})

	return resp, err
}

func (s *service) CreateUser(ctx context.Context, tx pgx.Tx, req *dto.RegisterRequest) (uuid.UUID, error) {

	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return uuid.Nil, apperr.Internal("failed to check email", err)
	}

	if exists {
		return uuid.Nil, apperr.Conflict("email already registered")
	}

	// Hash password
	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return uuid.Nil, apperr.Internal("failed to hash password", err)
	}

	// Create user
	userID, err := s.repo.CreateUser(ctx, tx, &User{
		Email:         req.Email,
		PasswordHash:  hash,
		EmailVerified: false,
		Status:        UserStatusActive,
	},
	)
	if err != nil {
		return uuid.Nil,
			apperr.Internal("failed to create user", err)
	}

	return userID, nil
}
func (s *service) SendVerification(ctx context.Context, tx pgx.Tx, userID uuid.UUID, email string) error {

	token, err := security.GenerateToken()
	if err != nil {
		return apperr.Internal("failed to generate verification token", err)
	}

	if err := s.repo.CreateEmailVerification(ctx, tx, userID, token, time.Now().Add(24*time.Hour)); err != nil {
		return apperr.Internal("failed to create verification record", err)
	}

	verifyURL := fmt.Sprintf(
		"http://localhost:8082/api/v1/auth/verify-email?token=%s",
		url.QueryEscape(token),
	)
	// TODO: send email
	fmt.Println("verification url:", verifyURL)

	return nil
}
func (s *service) EmailVerify(ctx context.Context, token string) error {

	verification, err := s.repo.GetVerificationByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NotFound("verification token not found")
		}
		return apperr.Internal("failed to get verification record", err)
	}

	if err := s.validateVerificationToken(verification, token); err != nil {
		return err
	}

	if err := s.repo.MarkEmailVerified(ctx, uuid.MustParse(verification.UserID)); err != nil {
		return apperr.Internal("failed to verify email", err)
	}

	return nil
}
func (s *service) validateVerificationToken(verification *EmailVerification, token string) error {

	if verification == nil {
		return apperr.NotFound("verification token not found")
	}

	if verification.Token != token {
		return apperr.BadRequest("invalid verification token")
	}

	if time.Now().After(verification.ExpiresAt) {
		return apperr.BadRequest("verification token expired")
	}

	if verification.VerifiedAt != nil {
		return apperr.BadRequest("email already verified")
	}

	return nil
}

func (s *service) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {

	// 1. Get user
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Unauthorized("invalid credentials")
		}
		return nil, apperr.Internal("failed to get user", err)
	}

	// 2. Email verification check
	if !user.EmailVerified {
		return nil, apperr.Forbidden("email not verified")
	}

	// 3. Check password
	ok, err := s.hasher.Verify(req.Password, user.PasswordHash)
	if err != nil {
		return nil, apperr.Internal("failed to verify password", err)
	}

	if !ok {
		return nil, apperr.Unauthorized("invalid credentials")
	}
	accessToken, err := s.jwt.CreateToken(user.ID, user.Email, s.jwt.Cfg.AccessTokenTTL)
	if err != nil {
		return nil, apperr.Internal("failed to generate access token", err)
	}
	// 5. Generate refresh token (random string)
	refreshToken, refreshHash, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, apperr.Internal("failed to generate refresh token", err)
	}

	// 6. Store session in DB
	err = s.repo.CreateSession(ctx, &Session{
		UserID:           uuid.MustParse(user.ID),
		RefreshTokenHash: refreshHash,
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		return nil, apperr.Internal("failed to create session", err)
	}

	// 7. Return response
	return &dto.LoginResponse{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) {

	// 1. Hash incoming refresh token
	hash := sha256.Sum256([]byte(refreshToken))
	refreshHash := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	// 2. Find session
	session, err := s.repo.GetSessionByRefreshHash(ctx, refreshHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Unauthorized("invalid refresh token")
		}
		return nil, apperr.Internal("failed to get session", err)
	}

	// 3. Check expired
	if time.Now().After(session.ExpiresAt) {
		return nil, apperr.Unauthorized("refresh token expired")
	}

	// 4. Check revoked
	if session.RevokedAt != nil {
		// reuse attack protection
		_ = s.repo.RevokeAll(ctx, session.UserID)

		return nil, apperr.Unauthorized("refresh token reused")
	}

	// 5. Revoke current session (ROTATION STEP 1)
	now := time.Now()
	if err := s.repo.RevokeSession(ctx, session.ID, now); err != nil {
		return nil, apperr.Internal("failed to revoke session", err)
	}

	// 6. Generate new access token
	user, err := s.repo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, apperr.Internal("failed to get user", err)
	}

	accessToken, err := s.jwt.CreateToken(user.ID, user.Email, s.jwt.Cfg.AccessTokenTTL)
	if err != nil {
		return nil, apperr.Internal("failed to create access token", err)
	}

	// 7. Generate NEW refresh token (ROTATION STEP 2)
	newRefreshToken, newHash, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, apperr.Internal("failed to generate refresh token", err)
	}

	// 8. Store NEW session
	err = s.repo.CreateSession(ctx, &Session{
		UserID:           session.UserID,
		RefreshTokenHash: newHash,
		ExpiresAt:        s.jwt.Cfg.RefreshTokenTTL,
	})
	if err != nil {
		return nil, apperr.Internal("failed to create session", err)
	}

	// 9. Response
	return &dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
