package tomb

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/deckhouse/deckhouse/dhctl/pkg/log"
)

var callbacks teardownCallbacks

func init() {
	callbacks = teardownCallbacks{
		waitCh:        make(chan struct{}, 1),
		interruptedCh: make(chan struct{}, 1),
	}
}

type callback struct {
	Name string
	Do   func()
}

type teardownCallbacks struct {
	mutex sync.RWMutex
	data  []callback

	exhausted        bool
	notInterruptable bool

	waitCh        chan struct{}
	interruptedCh chan struct{}
}

func (c *teardownCallbacks) registerOnShutdown(name string, cb func()) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = append(c.data, callback{Name: name, Do: cb})
	log.DebugF("teardown callback '%s' added, callbacks in queue: %d\n", name, len(c.data))
}

func (c *teardownCallbacks) shutdown() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Prevent double shutdown.
	if c.exhausted {
		return
	}

	log.DebugF("teardown started, queue length: %d\n", len(c.data))

	// Run callbacks in FIFO order to shutdown fundamental things last.
	for i := len(c.data) - 1; i >= 0; i-- {
		cb := c.data[i]
		log.DebugF("teardown callback %d: '%s' started\n", i, cb.Name)
		cb.Do()
		c.data[i] = callback{Name: "Stub", Do: func() {}}
		log.DebugF("teardown callback %d: '%s' done\n", i, cb.Name)
	}

	log.DebugLn("teardown is finished")
	c.exhausted = true
	close(c.waitCh)
}

func (c *teardownCallbacks) wait() {
	<-c.waitCh
}

func RegisterOnShutdown(process string, cb func()) {
	callbacks.registerOnShutdown(process, cb)
}

func Shutdown() {
	callbacks.shutdown()
}

func WaitShutdown() {
	callbacks.wait()
}

func IsInterrupted() bool {
	select {
	case <-callbacks.interruptedCh:
		return true
	default:
	}
	return false
}

func WithoutInterruptions(fn func()) {
	callbacks.notInterruptable = true
	defer func() { callbacks.notInterruptable = false }()
	fn()
}

func WaitForProcessInterruption() {
	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, syscall.SIGINT, syscall.SIGTERM)

Select:
	s := <-interruptCh

	switch s {
	case syscall.SIGTERM, syscall.SIGINT:
		if callbacks.notInterruptable {
			goto Select
		}

		// Wait for the second signal to kill the main process immediately.
		go func() {
			<-interruptCh
			log.ErrorLn("Killed by signal twice.")
			os.Exit(1)
		}()

		// Close interrupted channel to signal interruptable loops to stop.
		close(callbacks.interruptedCh)

		// Run all registered teardown callbacks and print an explanation at the end.
		callbacks.data = append([]callback{{
			Name: "Shutdown message",
			Do: func() {
				log.WarnLn(fmt.Sprintf("Graceful shutdown by %q signal ...", s.String()))
			},
		}}, callbacks.data...)
		Shutdown()
	default:
		os.Exit(1)
	}
}