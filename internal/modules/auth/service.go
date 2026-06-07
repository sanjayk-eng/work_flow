package auth

import (
	"context"
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
}

type service struct {
	repo        Repository
	hasher      *password.Hasher
	userservice user.Service
	posgressDb  *postgres.DB
}

func NewService(repo Repository, db *postgres.DB, hasher *password.Hasher, userservice user.Service) Service {
	return &service{repo: repo, posgressDb: db, hasher: hasher, userservice: userservice}
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

	// Check email already exists
	_, err := s.repo.GetByEmail(ctx, req.Email)
	switch {
	case err == nil:
		return uuid.Nil, apperr.Conflict("email already registered")

	case errors.Is(err, pgx.ErrNoRows):
		// continue

	default:
		return uuid.Nil, apperr.Internal("failed to check existing user", err)
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
		fmt.Println("err========", err.Error())
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
