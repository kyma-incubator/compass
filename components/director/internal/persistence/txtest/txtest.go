package txtest

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/stretchr/testify/mock"
)

func PersistenceContextThatExpectsCommit() *automock.PersistenceTx {

	persistTx := &automock.PersistenceTx{}
	persistTx.On("Commit").Return(nil).Once()
	return persistTx
}

func PersistenceContextThatDontExpectCommit() *automock.PersistenceTx {
	persistTx := &automock.PersistenceTx{}
	return persistTx
}

func TransactionerThatSucceed(persistTx *automock.PersistenceTx) *automock.Transactioner {
	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Once()
	transact.On("RollbackUnlessCommited", persistTx).Return().Once()
	return transact
}

func CtxWithDBMatcher() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		persistenceOp, err := persistence.FromCtx(ctx)
		return err == nil && persistenceOp != nil
	})
}
