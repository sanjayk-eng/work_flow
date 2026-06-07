package user

import (
	"context"
	"net/url"

	"github/sanjay-khandelwal/internal/modules/iam/user/dto"
	"github/sanjay-khandelwal/internal/shared/apperr"

	"github.com/jackc/pgx/v5"
)

type Service interface {
	CreateProfile(ctx context.Context, tx pgx.Tx, req *dto.CreateProfileRequest) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateProfile(ctx context.Context, tx pgx.Tx, req *dto.CreateProfileRequest) error {
	avatar := req.AvatarURL
	if avatar == "" {
		name := url.QueryEscape(req.FirstName + " " + req.LastName)
		avatar = "https://ui-avatars.com/api/?name=" + name
	}

	profile := &Profile{
		UserID:    req.UserID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AvatarURL: avatar,
	}

	if err := s.repo.CreateProfile(ctx, tx, profile); err != nil {
		return apperr.Internal("failed to create profile", err)
	}

	return nil
}
