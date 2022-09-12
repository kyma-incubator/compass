package txtest

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/stretchr/testify/mock"
)

// PersistenceContextThatExpectsCommit missing godoc
func PersistenceContextThatExpectsCommit() *automock.PersistenceTx {
	persistTx := &automock.PersistenceTx{}
	persistTx.On("Commit").Return(nil).Once()
	return persistTx
}

// PersistenceContextThatDoesntExpectCommit missing godoc
func PersistenceContextThatDoesntExpectCommit() *automock.PersistenceTx {
	persistTx := &automock.PersistenceTx{}
	return persistTx
}

// TransactionerThatSucceeds missing godoc
func TransactionerThatSucceeds(persistTx *automock.PersistenceTx) *automock.Transactioner {
	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Once()
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
	return transact
}

// TransactionerThatSucceedsTwice missing godoc
func TransactionerThatSucceedsTwice(persistTx *automock.PersistenceTx) *automock.Transactioner {
	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Twice()
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
	return transact
}

// TransactionerThatDoesARollback missing godoc
func TransactionerThatDoesARollback(persistTx *automock.PersistenceTx) *automock.Transactioner {
	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Once()
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
	return transact
}

// TransactionerThatDoesARollbackTwice missing godoc
func TransactionerThatDoesARollbackTwice(persistTx *automock.PersistenceTx) *automock.Transactioner {
	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Twice()
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Twice()
	return transact
}

// NoopTransactioner missing godoc
func NoopTransactioner(_ *automock.PersistenceTx) *automock.Transactioner {
	return &automock.Transactioner{}
}

// CtxWithDBMatcher missing godoc
func CtxWithDBMatcher() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		persistenceOp, err := persistence.FromCtx(ctx)
		return err == nil && persistenceOp != nil
	})
}

type txCtxGenerator struct {
	returnedError error
}

// NewTransactionContextGenerator missing godoc
func NewTransactionContextGenerator(potentialError error) *txCtxGenerator {
	return &txCtxGenerator{returnedError: potentialError}
}

// ThatSucceeds missing godoc
func (g txCtxGenerator) ThatSucceeds() (*automock.PersistenceTx, *automock.Transactioner) {
	return g.ThatSucceedsMultipleTimes(1)
}

func (g txCtxGenerator) ThatSucceedsTwice() (*automock.PersistenceTx, *automock.Transactioner) {
	return g.ThatSucceedsMultipleTimes(2)
}

// ThatSucceedsMultipleTimes missing godoc
func (g txCtxGenerator) ThatSucceedsMultipleTimes(times int) (*automock.PersistenceTx, *automock.Transactioner) {
	persistTx := &automock.PersistenceTx{}
	persistTx.On("Commit").Return(nil).Times(times)

	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Times(times)
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(times)

	return persistTx, transact
}

// ThatDoesntExpectCommit missing godoc
func (g txCtxGenerator) ThatDoesntExpectCommit() (*automock.PersistenceTx, *automock.Transactioner) {
	persistTx := &automock.PersistenceTx{}

	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Once()
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

	return persistTx, transact
}

// ThatFailsOnCommit missing godoc
func (g txCtxGenerator) ThatFailsOnCommit() (*automock.PersistenceTx, *automock.Transactioner) {
	persistTx := &automock.PersistenceTx{}
	persistTx.On("Commit").Return(g.returnedError).Once()

	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Once()
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

	return persistTx, transact
}

// ThatFailsOnBegin missing godoc
func (g txCtxGenerator) ThatFailsOnBegin() (*automock.PersistenceTx, *automock.Transactioner) {
	persistTx := &automock.PersistenceTx{}

	transact := &automock.Transactioner{}
	transact.On("Begin").Return(persistTx, g.returnedError).Once()

	return persistTx, transact
}

// ThatDoesntStartTransaction missing godoc
func (g txCtxGenerator) ThatDoesntStartTransaction() (*automock.PersistenceTx, *automock.Transactioner) {
	persistTx := &automock.PersistenceTx{}
	transact := &automock.Transactioner{}

	return persistTx, transact
}
