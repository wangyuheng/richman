package domain

type LedgerRepository interface {
	Save(it *Ledger) error
	QueryByUID(UID string) (*Ledger, bool)
}
