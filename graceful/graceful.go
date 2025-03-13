package graceful

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"
)

// GracefulShutdown предоставляет механизм для graceful shutdown
type Graceful struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
	timeout time.Duration
	errors  error
}

// Task представляет интерфейс для задач, которые могут быть запущены и остановлены
type GracefulTask interface {
	Run(context.Context) error
}

// NewGracefulShutdown создает новый экземпляр GracefulShutdown
func New(ctx context.Context, timeout time.Duration) *Graceful {
	ctx, cancel := context.WithCancel(ctx)
	return &Graceful{
		ctx:     ctx,
		cancel:  cancel,
		wg:      &sync.WaitGroup{},
		timeout: timeout,
	}
}

// AddTask добавляет задачи в Graceful
func (gs *Graceful) AddTask(tasks ...GracefulTask) {
	gs.wg.Add(len(tasks))

	for _, task := range tasks {
		go func() {
			defer gs.wg.Done()
			err := task.Run(gs.ctx)
			gs.errors = errors.Join(gs.errors, err)
		}()
	}
}

// Wait ожидает сигнала завершения и затем ожидает завершения всех задач
func (gs *Graceful) Wait(sig ...os.Signal) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, sig...)
	select {
	case <-gs.ctx.Done():
	case <-stop:
	}
	gs.cancel()

	// Создаем канал для отслеживания завершения задач
	done := make(chan struct{})
	go func() {
		gs.wg.Wait()
		close(done)
	}()

	// Ожидаем завершения задач или истечения времени ожидания
	select {
	case <-done:
		return gs.errors
	case <-time.After(gs.timeout):
		return gs.errors
	}
}
