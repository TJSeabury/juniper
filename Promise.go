package main

// A Result wraps a value and an error, similar to how a promise might resolve or reject
type Result[T any] struct {
	Value T
	Error error
}

// Promise represents the eventual completion (or failure) of an asynchronous operation and its resulting value.
type Promise[T any] struct {
	result chan Result[T]
}

// NewPromise initializes and returns a new Promise.
func NewPromise[T any](f func() (T, error)) *Promise[T] {
	p := &Promise[T]{result: make(chan Result[T], 1)} // Buffered channel to prevent blocking
	go func() {
		value, err := f()
		p.result <- Result[T]{Value: value, Error: err}
		close(p.result)
	}()
	return p
}

// Then registers callbacks for the fulfillment and rejection of the Promise.
func (p *Promise[T]) Then(successFn func(T), failureFn func(error)) {
	go func() {
		res := <-p.result
		if res.Error != nil {
			failureFn(res.Error)
		} else {
			successFn(res.Value)
		}
	}()
}

func (p *Promise[T]) Catch(failureFn func(error)) {
	go func() {
		res := <-p.result
		if res.Error != nil {
			failureFn(res.Error)
		}
	}()
}

func (p *Promise[T]) Finally(finallyFn func(T)) {
	go func() {
		res := <-p.result
		finallyFn(res.Value)
	}()
}

func All[T any](promises ...*Promise[T]) *Promise[[]T] {
	result := make([]T, len(promises))
	results := make([]Result[T], len(promises))
	for i, promise := range promises {
		promise.Then(func(value T) {
			result[i] = value
		}, func(err error) {
			results[i] = Result[T]{Error: err}
		})
	}
	return NewPromise(func() ([]T, error) {
		for _, result := range results {
			if result.Error != nil {
				return nil, result.Error
			}
		}
		return result, nil
	})
}
