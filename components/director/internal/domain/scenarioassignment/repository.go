package scenarioassignment

import "github.com/kyma-incubator/compass/components/director/internal/repo"

func NewRepository() *repository {
	return &repository{
		creator: repo.NewCreator()
	}
}

type repository struct {
	creator repo.Creator
}
