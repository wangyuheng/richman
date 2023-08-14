package database

import "context"

type Cacheable interface {
	WarmUP(ctx context.Context)
	Key(s string) string
}
