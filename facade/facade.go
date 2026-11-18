// Package facade holds the technical contracts the application layer depends on (the kimoj-core facades):
// transactions with master/slave routing, event bus, cache, password hashing, id generation.
package facade

import (
	"context"
	"time"

	"github.com/solomanpan/kimoj-core/exception"
	"github.com/solomanpan/kimoj-core/fp"
)

// ---------------------------------------------------------------- datasource + transactions
type DataSource string

const (
	Master DataSource = "master" // read-write
	Slave  DataSource = "slave"  // read-only replica
)

type Tx interface {
	DataSource() DataSource
	// Root is false when this scope joined an outer one (only the root commits / rolls back).
	Root() bool
	SetRollbackOnly()
	// AfterCommit runs after the ROOT transaction commits (never on rollback): used to publish domain events.
	AfterCommit(fn func(ctx context.Context))
}

type txKey struct{}

func WithTx(ctx context.Context, tx Tx) context.Context { return context.WithValue(ctx, txKey{}, tx) }

// Current is the transaction of this call chain, if any.
func Current(ctx context.Context) (Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(Tx)
	return tx, ok
}

type TransactionManager interface {
	// Run executes fn inside a scope on the datasource. A panic rolls back and propagates.
	Run(ctx context.Context, ds DataSource, fn func(ctx context.Context))
}

// InMaster is @master: run fn in a read-write transaction (joins an outer one).
func InMaster[T any](m TransactionManager, ctx context.Context, fn func(context.Context) T) T {
	var out T
	m.Run(ctx, Master, func(c context.Context) { out = fn(c) })
	return out
}

// InSlave is @slave: run fn on the read replica (joins an outer scope).
func InSlave[T any](m TransactionManager, ctx context.Context, fn func(context.Context) T) T {
	var out T
	m.Run(ctx, Slave, func(c context.Context) { out = fn(c) })
	return out
}

// InMasterResult is @master for business results: a failed Result rolls the transaction back.
func InMasterResult[A any](m TransactionManager, ctx context.Context, fn func(context.Context) fp.Result[A]) fp.Result[A] {
	return InMaster(m, ctx, func(c context.Context) fp.Result[A] {
		res := fn(c)
		if fp.IsErr(res) {
			if tx, ok := Current(c); ok {
				tx.SetRollbackOnly()
			}
		}
		return res
	})
}

// ---------------------------------------------------------------- events
type Event interface {
	EventID() string
	EventType() string
	OrgID() string
	// AggregateKey orders events of one aggregate (Kafka key / stream field).
	AggregateKey() string
}

type Publisher interface {
	Publish(ctx context.Context, e Event) error
}

type Handler func(ctx context.Context, e Event) error

type Lifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type EventBus interface {
	Publisher
	Lifecycle
	// Subscribe registers a handler for one event type (before Start).
	Subscribe(eventType string, h Handler)
}

// ---------------------------------------------------------------- cache / security / ids
type Cache interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key, value string, ttl time.Duration)
	Delete(ctx context.Context, keys ...string)
	Close() error
}

type PasswordHasher interface {
	Hash(raw string) string
	Verify(raw, hashed string) bool
}

type IDGenerator func(ctx context.Context) string

// MustPublish raises an InfrastructureError when the broker rejects the event.
func MustPublish(ctx context.Context, p Publisher, e Event) {
	exception.Must(p.Publish(ctx, e), "publish_failed", "publish "+e.EventType())
}
