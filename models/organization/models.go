package organization

import (
	"database/sql"
	"github.com/pennsieve/pennsieve-go-api/models/dbTable"
	"log"
)

type Claim struct {
	Role            dbTable.DbPermission
	IntId           int64
	EnabledFeatures []dbTable.FeatureFlags
}

func GetOrganizationClaim(db *sql.DB, userId int64, organizationId int64) (*Claim, error) {

	var orgUser dbTable.OrganizationUser
	currentOrgUser, err := orgUser.GetByUserId(db, userId)
	if err != nil {
		log.Println("Unable to check Org User: ", err)
		return nil, err
	}

	var orgFeat dbTable.FeatureFlags
	allFeatures, err := orgFeat.GetAll(db, organizationId)
	if err != nil {
		log.Println("Unable to check Feature Flags: ", err)
		return nil, err
	}

	orgRole := Claim{
		Role:            currentOrgUser.DbPermission,
		IntId:           organizationId,
		EnabledFeatures: allFeatures,
	}

	return &orgRole, nil

}
