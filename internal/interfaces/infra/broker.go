package infra

import (
	"context"

	"github.com/sunr3d/simple-http-calendar/models"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.2 --name=Broker --output=../../../mocks --filename=mock_broker.go --with-expecter
type Broker interface {
	Publish(ctx context.Context, event *models.Event) error
	Subscribe(ctx context.Context, handler func(ctx context.Context, event *models.Event) error) error
}
