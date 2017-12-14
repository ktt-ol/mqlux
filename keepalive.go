package mqlux

import (
	"os"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

type watchdogHandler struct {
	killAfterSilence time.Duration
	keepAlive        chan struct{}
	done             chan struct{}
}

func NewWatchdogHandler(killAfterSilence time.Duration) *watchdogHandler {
	w := watchdogHandler{
		killAfterSilence: killAfterSilence,
		keepAlive:        make(chan struct{}),
		done:             make(chan struct{}),
	}
	go w.run()
	return &w
}
func (w *watchdogHandler) Receive(client mqtt.Client, message mqtt.Message) {
	select {
	case w.keepAlive <- struct{}{}:
	default:
		// ignore if keepAlive is closed, can happen when Watchdog was stopped
		// (there is no way to unsubscribe).
	}
}

func (w *watchdogHandler) Topic() string {
	return "/#"
}

func (w *watchdogHandler) Match(topic string) bool {
	return true
}

func (w *watchdogHandler) Stop() {
	w.done <- struct{}{}
}

func (w *watchdogHandler) run() {
	t := time.NewTicker(10 * time.Second)
	lastKeepAlive := time.Now()
	for {
		select {
		case <-t.C:
			if time.Since(lastKeepAlive) > w.killAfterSilence {
				os.Exit(42)
			}
		case <-w.keepAlive:
			lastKeepAlive = time.Now()
		case <-w.done:
			t.Stop()
			return
		}
	}
}
