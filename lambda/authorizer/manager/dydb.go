package manager

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dydb"
)

// PennsieveDyAPI is an interface only containing the methods of *dydb.Queries that are used by the ClaimsManager.
type PennsieveDyAPI interface {
	GetManifestById(ctx context.Context, manifestTableName string, manifestId string) (*dydb.ManifestTable, error)
}
