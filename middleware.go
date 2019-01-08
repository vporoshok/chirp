package chirp

import (
	"context"
	"log"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
)

type middlewareConfig struct {
	logger    *log.Logger
	interrupt int
	onerror   func(*Error, http.ResponseWriter, *http.Request) bool
}

func (cfg middlewareConfig) HandleError(err *Error, w http.ResponseWriter, req *http.Request) bool {
	if cfg.logger != nil {
		cfg.logger.Printf("request parsing error: %s", err.Error())
	}
	if cfg.onerror != nil {
		if cfg.onerror(err, w, req) {

			return true
		}
	}
	if cfg.interrupt != 0 {
		http.Error(w, err.Error(), cfg.interrupt)

		return true
	}

	return false
}

// Option configure Middleware
type Option interface {
	apply(cfg *middlewareConfig)
}

type optionFunc func(cfg *middlewareConfig)

func (fn optionFunc) apply(cfg *middlewareConfig) {
	fn(cfg)
}

// WithLogger log parse request errors
func WithLogger(logger *log.Logger) Option {

	return optionFunc(func(cfg *middlewareConfig) {
		cfg.logger = logger
	})
}

// WithInterrupt interrupt on error with given status
func WithInterrupt(status int) Option {

	return optionFunc(func(cfg *middlewareConfig) {
		cfg.interrupt = status
	})
}

// WithOnError add an error handler
//
// If handler return true, request might be interrupt immediate.
func WithOnError(cb func(*Error, http.ResponseWriter, *http.Request) bool) Option {

	return optionFunc(func(cfg *middlewareConfig) {
		cfg.onerror = cb
	})
}

type ctxKey struct{}

// Middleware to inject request struct
//
// For every request middleware create new instance of dst struct and parse
// request to it. Result will be added to request context as non-pointered
// value.
func Middleware(dst interface{}, options ...Option) func(next http.Handler) http.Handler {
	cfg := middlewareConfig{}
	for _, opt := range options {
		opt.apply(&cfg)
	}

	dstType := reflect.TypeOf(dst)
	if dstType.Kind() != reflect.Struct {

		panic(errors.New("dst should be not pointered struct"))
	}

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			value := reflect.New(dstType)
			err := parse(req, value)
			result := value.Elem().Interface()
			req = req.WithContext(context.WithValue(req.Context(), ctxKey{}, result))
			if err != nil {
				if cfg.HandleError(err, w, req) {

					return
				}
			}
			next.ServeHTTP(w, req)
		})
	}
}

// RequestFromContext extract request struct instance from context
func RequestFromContext(ctx context.Context) interface{} {

	return ctx.Value(ctxKey{})
}
