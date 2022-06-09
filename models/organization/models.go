package organization

import (
	"github.com/pennsieve/pennsieve-go-api/models/dbTable"
)

// Claim combines the role of the user in the org, and the features in the organization.
type Claim struct {
	Role            dbTable.DbPermission
	IntId           int64
	EnabledFeatures []dbTable.FeatureFlags
}
