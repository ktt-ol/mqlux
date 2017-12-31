package router

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/ktt-ol/mqlux/internal/mqlux"
)

type Receiver interface {
	// Receive takes and processes an incoming Message.
	Receive(message mqlux.Message)
}

// New returns a new Router for fast MQTT topic look ups.
func New() *Router {
	return &Router{}
}

type Router struct {
	topics []handler
	sorted bool
	mu     sync.RWMutex
}

// Add adds a new path to the router and assigns it to a handler.
// The path can end with # to match any sub-path.
// The + wildcard is not supported.
func (r *Router) Add(topic string, h Receiver) {
	path := strings.Split(topic, "/")
	r.mu.Lock()
	r.sorted = false
	r.topics = append(r.topics, handler{path: path, Receiver: h})
	r.mu.Unlock()
}

func (r *Router) Find(topic string) []Receiver {
	path := strings.Split(topic, "/")
	r.mu.RLock()
	defer r.mu.RUnlock()

	for !r.sorted {
		// need to sort r.topic, upgrade read lock to lock
		// uses for !r.sorted to prevent race condition
		r.mu.RUnlock()
		r.mu.Lock()
		sort.Sort(byPath(r.topics))
		r.sorted = true
		r.mu.Unlock()
		r.mu.RLock()
	}

	result := r.find(path, nil)
	wildcard := make([]string, len(path)+1)
	copy(wildcard, path)
	for i := len(path); i >= 0; i-- {
		wildcard[i] = "#"
		wildcard = wildcard[:i+1]
		result = r.find(wildcard, result)
	}
	return result
}

func (r *Router) find(path []string, result []Receiver) []Receiver {
	i := sort.Search(len(r.topics), func(i int) bool {
		for j, p := range path {
			if len(r.topics[i].path) < j+1 {
				return false
			}
			if r.topics[i].path[j] < p {
				return false
			}
		}
		return true
	})
	if i >= 0 && i < len(r.topics) {
		for j := i; j < len(r.topics); j++ {
			if identical(r.topics[j].path, path) {
				result = append(result, r.topics[j].Receiver)
			} else {
				break
			}
		}
	}
	return result
}

func (r *Router) Receive(msg mqlux.Message) {
	handlers := r.Find(msg.Topic)
	// log.Printf("debug: forwarding %s to %d handlers", msg.Topic, len(handlers))
	for _, h := range handlers {
		h.Receive(msg)
	}
}

// hasPrefix checks whether paths starts with prefix
func hasPrefix(path, prefix []string) bool {
	fmt.Println(path, prefix)
	for i := range prefix {
		if len(path) < i+1 {
			return false
		}
		if path[i] != prefix[i] {
			return false
		}
	}
	return true
}

// identical checks whether both paths are identical
func identical(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type handler struct {
	path []string
	Receiver
}

type byPath []handler

func (p byPath) Len() int      { return len(p) }
func (p byPath) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p byPath) Less(i, j int) bool {
	ap := p[i].path
	bp := p[j].path
	for i := range ap {
		if len(bp) < (i + 1) {
			return false
		}
		if ap[i] < bp[i] {
			return true
		} else if ap[i] > bp[i] {
			return false
		}
	}
	return true
}
