package master

import "context"

type Service interface {
	Start(ctx context.Context)
	Stop()
}
