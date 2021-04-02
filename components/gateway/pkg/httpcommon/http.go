package httpcommon

import (
	"context"
	"io"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

func CloseBody(ctx context.Context, body io.ReadCloser) {
	if err := body.Close(); err != nil {
		log.C(ctx).WithError(err).Errorf("while closing body: %v", err)
	}
}
