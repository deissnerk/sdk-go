package pipeline

import (
	"context"
)

type StartStop interface {
	Start() error
	Stop(ctx context.Context)
}
