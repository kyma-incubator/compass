package token

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/kyma-incubator/compass/components/connector/pkg/repo"
)

var (
	columnNames = []string{"id", "token", "token_type", "client_id", "created_at", "used_at", "used"}
)

const (
	tableName = "public.token"
)

type Repository struct {
	creator repo.Creator
	getter  repo.SingleGetterGlobal
	updater repo.UpdaterGlobal
}

func NewRepository() *Repository {
	return &Repository{
		creator: repo.NewCreator(repo.Token, tableName, columnNames),
		getter:  repo.NewSingleGetterGlobal(repo.Token, tableName, columnNames),
		updater: repo.NewUpdaterGlobal(repo.Token, tableName, []string{"used", "used_at"}, []string{"token"}),
	}
}

func (r *Repository) Create(ctx context.Context, token string, tokenData tokens.TokenData) error {
	return r.creator.Create(ctx, &Entity{
		ID:        uuid.New().String(),
		Token:     token,
		TokenType: string(tokenData.Type),
		ClientID:  tokenData.ClientId,
		CreatedAt: time.Now(),
		UsedAt:    time.Time{},
		Used:      false,
	})
}

// TODO: Introduce app model layer?
func (r *Repository) Get(ctx context.Context, token string) (tokens.Token, apperrors.AppError) {
	var entity Entity
	// TODO: Handle errors
	if err := r.getter.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("token", token), repo.NewEqualCondition("used", false)}, repo.NoOrderBy, &entity); err != nil {
		return tokens.Token{}, apperrors.Internal("while getting token %s", err)
	}
	return tokens.Token{
		TokenData: tokens.TokenData{
			Type:     tokens.TokenType(entity.TokenType),
			ClientId: entity.ClientID,
		},
		CreatedAt: entity.CreatedAt,
		Used:      entity.Used,
	}, nil
}

func (r *Repository) Invalidate(ctx context.Context, token string) apperrors.AppError {
	// TODO: Handle errors
	if err := r.updater.UpdateSingleGlobal(ctx, &Entity{
		Token:  token,
		UsedAt: time.Now(),
		Used:   true,
	}); err != nil {
		return apperrors.Internal("while invalidating token %s", err)
	}
	return nil
}
