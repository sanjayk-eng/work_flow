package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github/sanjay-khandelwal/internal/modules/iam/auth/dto"
	"github/sanjay-khandelwal/internal/modules/iam/session"
	"github/sanjay-khandelwal/internal/modules/iam/user"
	userDto "github/sanjay-khandelwal/internal/modules/iam/user/dto"
	"github/sanjay-khandelwal/internal/shared/apperr"
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"
	"github/sanjay-khandelwal/internal/shared/security"
	"github/sanjay-khandelwal/internal/shared/security/jwt"
	"github/sanjay-khandelwal/internal/shared/security/password"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error)
	EmailVerify(ctx context.Context, token string) error
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error)

	// Logout revokes the session associated with the given refresh token.
	Logout(ctx context.Context, refreshToken string) error

	// LogoutAll revokes every active session for the given user.
	LogoutAll(ctx context.Context, userID uuid.UUID) error
}

type service struct {
	repo        Repository
	hasher      *password.Hasher
	jwt         *jwt.Service
	userService user.Service
	sessionSvc  session.Service
	db          *postgres.DB
}

func NewService(
	repo Repository,
	db *postgres.DB,
	hasher *password.Hasher,
	jwtSvc *jwt.Service,
	userService user.Service,
	sessionSvc session.Service,
) Service {
	return &service{
		repo:        repo,
		db:          db,
		hasher:      hasher,
		jwt:         jwtSvc,
		userService: userService,
		sessionSvc:  sessionSvc,
	}
}

func (s *service) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	var resp *dto.UserResponse

	err := s.db.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		userID, err := s.CreateUser(ctx, tx, req)
		if err != nil {
			return err
		}

		if err := s.userService.CreateProfile(ctx, tx, &userDto.CreateProfileRequest{
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

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return uuid.Nil, apperr.Internal("failed to hash password", err)
	}

	userID, err := s.repo.CreateUser(ctx, tx, &User{
		Email:         req.Email,
		PasswordHash:  hash,
		EmailVerified: false,
		Status:        UserStatusActive,
	})
	if err != nil {
		return uuid.Nil, apperr.Internal("failed to create user", err)
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
		"http://localhost:5173/auth/verify-email?token=%s",
		url.QueryEscape(token),
	)
	// TODO: replace with email service
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

func (s *service) validateVerificationToken(v *EmailVerification, token string) error {
	if v == nil {
		return apperr.NotFound("verification token not found")
	}
	if v.Token != token {
		return apperr.BadRequest("invalid verification token")
	}
	if time.Now().After(v.ExpiresAt) {
		return apperr.BadRequest("verification token expired")
	}
	if v.VerifiedAt != nil {
		return apperr.BadRequest("email already verified")
	}
	return nil
}

func (s *service) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	u, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Unauthorized("invalid credentials")
		}
		return nil, apperr.Internal("failed to get user", err)
	}

	if !u.EmailVerified {
		return nil, apperr.Forbidden("email not verified")
	}

	ok, err := s.hasher.Verify(req.Password, u.PasswordHash)
	if err != nil {
		fmt.Println(err.Error())
		return nil, apperr.Internal("failed to verify password", err)
	}
	if !ok {
		return nil, apperr.Unauthorized("invalid credentials")
	}

	accessToken, err := s.jwt.CreateToken(u.ID, u.Email, s.jwt.Cfg.AccessTokenTTL)
	if err != nil {
		return nil, apperr.Internal("failed to generate access token", err)
	}

	refreshToken, refreshHash, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, apperr.Internal("failed to generate refresh token", err)
	}

	if err := s.sessionSvc.Create(ctx, uuid.MustParse(u.ID), refreshHash, time.Now().Add(7*24*time.Hour)); err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		UserID:       u.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) {
	// 1. Hash incoming token to look up session
	hash := sha256.Sum256([]byte(refreshToken))
	refreshHash := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	// 2. Find session
	sess, err := s.sessionSvc.GetByRefreshHash(ctx, refreshHash)
	if err != nil {
		return nil, err
	}

	// 3. Expired check
	if time.Now().After(sess.ExpiresAt) {
		return nil, apperr.Unauthorized("refresh token expired")
	}

	// 4. Revoked check — reuse-attack protection
	if sess.RevokedAt != nil {
		_ = s.sessionSvc.RevokeAll(ctx, sess.UserID)
		return nil, apperr.Unauthorized("refresh token reused")
	}

	// 5. Rotate: revoke current session
	if err := s.sessionSvc.Revoke(ctx, sess.ID); err != nil {
		return nil, err
	}

	// 6. Load user for new token claims
	u, err := s.repo.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, apperr.Internal("failed to get user", err)
	}

	accessToken, err := s.jwt.CreateToken(u.ID, u.Email, s.jwt.Cfg.AccessTokenTTL)
	if err != nil {
		return nil, apperr.Internal("failed to create access token", err)
	}

	// 7. Issue new refresh token (rotation)
	newRefreshToken, newHash, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, apperr.Internal("failed to generate refresh token", err)
	}

	if err := s.sessionSvc.Create(ctx, sess.UserID, newHash, s.jwt.Cfg.RefreshTokenTTL); err != nil {
		return nil, err
	}

	return &dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	// Hash the incoming token to locate the session row
	hash := sha256.Sum256([]byte(refreshToken))
	refreshHash := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	sess, err := s.sessionSvc.GetByRefreshHash(ctx, refreshHash)
	if err != nil {
		// Token not found or already revoked — treat as success (idempotent logout)
		return nil
	}

	return s.sessionSvc.Revoke(ctx, sess.ID)
}

func (s *service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.sessionSvc.RevokeAll(ctx, userID)
}
