package application

import (
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sqlx/dto"
	"testing"
)

func TestAbd(t *testing.T) {
	appDTO := dto.ApplicationDTO{}
	err := sq.Insert("applications").Scan(appDTO).Error()
	fmt.Println("err",err)
}
