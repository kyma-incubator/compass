package automock

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/stretchr/testify/mock"
)

func PersistenceContextThatExpectsCommit() *PersistenceTx {
	persistTx := &PersistenceTx{}
	persistTx.On("Commit").Return(nil).Once()
	return persistTx
}

func PersistenceContextThatDontExpectCommit() *PersistenceTx {
	persistTx := &PersistenceTx{}
	return persistTx
}

func TransactionerThatSucceed(persistTx *PersistenceTx) *Transactioner {
	transact := &Transactioner{}
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
