// Package fp is the thin functional layer over IBM/fp-go: Result is Either[DomainError, A].
package fp

import (
	"github.com/IBM/fp-go/v2/either"
	"github.com/IBM/fp-go/v2/option"

	"github.com/solomanpan/kimoj-core/exception"
)

type Result[A any] = either.Either[exception.DomainError, A]
type Option[A any] = option.Option[A]

func Ok[A any](a A) Result[A] { return either.Right[exception.DomainError](a) }

func Err[A any](e exception.DomainError) Result[A] { return either.Left[A](e) }

func Bind[A, B any](r Result[A], f func(A) Result[B]) Result[B] { return either.MonadChain(r, f) }

func Map[A, B any](r Result[A], f func(A) B) Result[B] { return either.MonadMap(r, f) }

// Then keeps the value of r when next succeeds (like Python's AsyncResult.ensure).
func Then[A, B any](r Result[A], next func(A) Result[B]) Result[A] {
	return either.MonadChain(r, func(a A) Result[A] {
		return either.MonadMap(next(a), func(B) A { return a })
	})
}

// Get unpacks a Result: (value, nil) or (zero, &error).
func Get[A any](r Result[A]) (A, *exception.DomainError) {
	a, e := either.Unwrap(r)
	if either.IsLeft(r) {
		return a, &e
	}
	return a, nil
}

func IsErr[A any](r Result[A]) bool { return either.IsLeft(r) }

func Fold[A, B any](r Result[A], onErr func(exception.DomainError) B, onOk func(A) B) B {
	return either.Fold(onErr, onOk)(r)
}

// Some / None / Find: optional values (a repository lookup that may find nothing).
func Some[A any](a A) Option[A] { return option.Some(a) }

func None[A any]() Option[A] { return option.None[A]() }

func IsSome[A any](o Option[A]) bool { return option.IsSome(o) }

// Value returns the contained value and whether there was one.
func Value[A any](o Option[A]) (A, bool) {
	var zero A
	return option.GetOrElse(func() A { return zero })(o), option.IsSome(o)
}

// OrNotFound turns an Option into a Result using the generic not-found error.
func OrNotFound[A any](o Option[A], entity, id string) Result[A] {
	if v, ok := Value(o); ok {
		return Ok(v)
	}
	return Err[A](exception.NewNotFound(entity, id))
}
