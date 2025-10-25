package services

import (
	"context"
)

type ArchiveService interface {
	Start(ctx context.Context) error
}
