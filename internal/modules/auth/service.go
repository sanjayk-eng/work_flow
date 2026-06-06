package auth

import (
	"context"
	"time"

	"github/sanjay-khandelwal/internal/modules/auth/dto"
	"github/sanjay-khandelwal/internal/shared/apperr"
	"github/sanjay-khandelwal/internal/shared/security/password"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Service is the interface handlers depend on.
type Service interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error)
}

type service struct {
	repo   Repository
	hasher *password.Hasher
}

func NewService(repo Repository, hasher *password.Hasher) Service {
	return &service{repo: repo, hasher: hasher}
}

// Register creates a new user + profile inside a single transaction.
//
// Flow:
//  1. Check email uniqueness
//  2. Hash password (Argon2id)
//  3. Begin transaction
//  4. Insert user row
//  5. Insert profile row
//  6. Commit → return response
func (s *service) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// 1. Check email uniqueness
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperr.Conflict("email already in use")
	}

	// 2. Hash password
	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, apperr.Internal("failed to hash password", err)
	}

	now := time.Now().UTC()
	userID := uuid.New().String()

	user := &User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: hash,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	profile := &Profile{
		UserID:    userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 3-5. Insert user + profile atomically
	if err := s.repo.WithTx(ctx, func(txCtx context.Context, tx pgx.Tx) error {
		if err := s.repo.CreateUser(txCtx, tx, user); err != nil {
			return err
		}
		return s.repo.CreateProfile(txCtx, tx, profile)
	}); err != nil {
		return nil, err
	}

	// 6. Return response — mapping happens inside dto, service just passes data
	return &dto.RegisterResponse{
		User: dto.NewUserResponse(dto.UserResponseInput{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: profile.FirstName,
			LastName:  profile.LastName,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		}),
	}, nil
}
