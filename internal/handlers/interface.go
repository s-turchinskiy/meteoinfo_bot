package handlers

import "context"

type Handlerer interface {
	Do(ctx context.Context)
	Close(ctx context.Context) error
}
