package domain

import "context"

type LedgerRepository interface {
	Save(it *Ledger) error
	QueryByUID(UID string) (*Ledger, bool)
	QueryUnallocated() []*Ledger
	UpdateUser(id string, user User) error
	WarmUP(ctx context.Context)
}
