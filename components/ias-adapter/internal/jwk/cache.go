package jwk

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
)

type Cache struct {
	jwksEndpoint string
	cache        *jwk.Cache
}

func NewJWKCache(ctx context.Context, cfg config.JWKCache) (Cache, error) {
	cache := jwk.NewCache(ctx, jwk.WithRefreshWindow(cfg.SyncInterval))
	cache.Register(cfg.Endpoint)
	if _, err := cache.Refresh(ctx, cfg.Endpoint); err != nil {
		return Cache{}, errors.Newf("failed to refresh jwks: %w", err)
	}
	return Cache{
		jwksEndpoint: cfg.Endpoint,
		cache:        cache,
	}, nil
}

func (c Cache) Get(ctx context.Context, kid string) (any, error) {
	set, err := c.cache.Get(ctx, c.jwksEndpoint)
	if err != nil {
		return nil, errors.Newf("failed to get jwks set: %w", err)
	}

	for it := set.Iterate(context.Background()); it.Next(context.Background()); {
		key := it.Pair().Value.(jwk.Key)
		if key.KeyID() == kid {
			alg := key.Algorithm().String()
			fmt.Println("########", alg)
			if alg != jwt.SigningMethodRS256.Alg() {
				return nil, errors.Newf("unexpected jwk algorithm '%s'", alg)
			}
			var rawkey any
			return rawkey, key.Raw(&rawkey)
		}
	}

	return nil, errors.New("jwk not found")
}
