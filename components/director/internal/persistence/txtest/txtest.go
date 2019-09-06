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

func PersistenceContextThatDoesntExpectCommit() *automock.PersistenceTx {
	persistTx := &automock.PersistenceTx{}
	return persistTx
}

func TransactionerThatSucceeds(persistTx *automock.PersistenceTx) *automock.Transactioner {
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

type txCtxGenerator struct {
	returnedError error
}

func NewTransactionContextGenerator(potentialError error) *txCtxGenerator {
	return &txCtxGenerator{returnedError: potentialError}
}

func (g txCtxGenerator) ThatSucceeds() (*automock.PersistenceTx, *automock.Transactioner) {
	persistTx := &automock.PersistenceTx{}
	persistTx.On("Commit").Return(nil).Once()

	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Once()
	transact.On("RollbackUnlessCommited", persistTx).Return().Once()

	return persistTx, transact
}

func (g txCtxGenerator) ThatDoesntExpectCommit() (*automock.PersistenceTx, *automock.Transactioner) {
	persistTx := &automock.PersistenceTx{}

	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Once()
	transact.On("RollbackUnlessCommited", persistTx).Return().Once()

	return persistTx, transact
}

func (g txCtxGenerator) ThatFailsOnCommit() (*automock.PersistenceTx, *automock.Transactioner) {
	persistTx := &automock.PersistenceTx{}
	persistTx.On("Commit").Return(g.returnedError).Once()

	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Once()
	transact.On("RollbackUnlessCommited", persistTx).Return().Once()

	return persistTx, transact
}

func (g txCtxGenerator) ThatFailsOnBegin() (*automock.PersistenceTx, *automock.Transactioner) {
	persistTx := &automock.PersistenceTx{}

	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, g.returnedError).Once()

	return persistTx, transact
}

func (g txCtxGenerator) ThatDoesntStartTransaction() (*automock.PersistenceTx, *automock.Transactioner) {
	persistTx := &automock.PersistenceTx{}
	transact := &automock.Transactioner{}

	return persistTx, transact
}
