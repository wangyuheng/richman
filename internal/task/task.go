package task

import (
	"context"
	"github.com/wangyuheng/richman/internal/domain"
	"github.com/wangyuheng/richman/internal/usecase"
	"log"
	"time"
)

type Tasker interface {
	Start()
}

type WarmTask struct {
	timer      *time.Ticker
	warmTimer  *time.Ticker
	ledger     usecase.LedgerUseCase
	ledgerRepo domain.LedgerRepository
	userRepo   domain.UserRepository
}

func (t *WarmTask) Start() {
	go func() {
		for range t.timer.C {
			log.Println("定时执行 PreparedAllocated")
			_ = t.ledger.PreparedAllocated()
		}
	}()
	go func() {
		for range t.warmTimer.C {
			ctx := context.Background()
			log.Println("定时执行 ledger WarmUP")
			t.ledgerRepo.WarmUP(ctx)
			log.Println("定时执行 user WarmUP")
			t.userRepo.WarmUP(ctx)
		}
	}()
}

func NewWarmTask(ledger usecase.LedgerUseCase, userRepo domain.UserRepository, ledgerRepo domain.LedgerRepository) Tasker {
	return &WarmTask{
		timer:      time.NewTicker(5 * time.Minute),
		warmTimer:  time.NewTicker(1 * time.Hour),
		ledger:     ledger,
		ledgerRepo: ledgerRepo,
		userRepo:   userRepo,
	}
}
