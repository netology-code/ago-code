package remux

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var paramsContextKey = &contextKey{"remux context"}

type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return c.name
}

type Params struct {
	Named      map[string]string
	Positional []string
}

type ReMux struct {
	mu              sync.RWMutex
	plain           map[Method]map[string]http.Handler
	regex           map[Method]map[*regexp.Regexp]http.Handler
	notFoundHandler http.Handler
}

var defaultNotFoundHandler = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
}

func NewReMux() *ReMux {
	return &ReMux{
		notFoundHandler: http.HandlerFunc(defaultNotFoundHandler),
	}
}

var (
	ErrInvalidPath      = errors.New("invalid path")
	ErrInvalidMethod    = errors.New("invalid http method")
	ErrNilHandler       = errors.New("nil handler")
	ErrAmbiguousMapping = errors.New("ambiguous mapping")
	ErrNoParams         = errors.New("no params")
)

type Method string

const (
	GET     Method = "GET"
	POST    Method = "POST"
	PUT     Method = "PUT"
	PATCH   Method = "PATCH"
	DELETE  Method = "DELETE"
	OPTIONS Method = "OPTIONS"
	HEAD    Method = "HEAD"
)

type Middleware func(handler http.Handler) http.Handler

func (r *ReMux) RegisterPlain(
	method Method,
	path string,
	handler http.Handler,
	middlewares ...Middleware,
) error {
	if !isValidMethod(method) {
		return ErrInvalidMethod
	}

	if !strings.HasPrefix(path, "/") {
		return ErrInvalidPath
	}

	if handler == nil {
		return ErrNilHandler
	}

	handler = wrapHandler(handler, middlewares...)

	r.mu.Lock()
	defer r.mu.Unlock()

	// запрещаем добавлять дубликаты
	if _, exists := r.plain[method][path]; exists {
		return ErrAmbiguousMapping
	}

	if r.plain == nil {
		r.plain = make(map[Method]map[string]http.Handler)
	}

	if r.plain[method] == nil {
		r.plain[method] = make(map[string]http.Handler)
	}

	r.plain[method][path] = handler
	return nil
}

func (r *ReMux) RegisterRegex(
	method Method,
	path *regexp.Regexp,
	handler http.Handler,
	middlewares ...Middleware,
) error {
	if !isValidMethod(method) {
		return ErrInvalidMethod
	}

	if !strings.HasPrefix(path.String(), `^/`) {
		return ErrInvalidPath
	}

	if !strings.HasSuffix(path.String(), `$`) {
		return ErrInvalidPath
	}

	if handler == nil {
		return ErrNilHandler
	}

	handler = wrapHandler(handler, middlewares...)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.regex[method][path]; exists {
		return ErrAmbiguousMapping
	}

	if r.regex == nil {
		r.regex = make(map[Method]map[*regexp.Regexp]http.Handler)
	}

	if r.regex[method] == nil {
		r.regex[method] = make(map[*regexp.Regexp]http.Handler)
	}

	r.regex[method][path] = handler
	return nil
}

func (r *ReMux) NotFound(handler http.Handler) error {
	if handler == nil {
		return ErrNilHandler
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notFoundHandler = handler
	return nil
}

func (r *ReMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.mu.RLock()
	var resultHandler http.Handler
	if handlers, exists := r.plain[Method(request.Method)]; exists {
		if handler, ok := handlers[request.URL.Path]; ok {
			resultHandler = handler
		}
	}
	if resultHandler == nil {
		if handlers, exists := r.regex[Method(request.Method)]; exists {
			for path, handler := range handlers {
				if matches := path.FindStringSubmatch(request.URL.Path); matches != nil {
					params := &Params{
						Positional: matches[1:],
						Named:      make(map[string]string),
					}
					for index, name := range path.SubexpNames() {
						if name == "" {
							continue
						}
						params.Named[name] = matches[index]
					}

					ctx := context.WithValue(request.Context(), paramsContextKey, params)

					request = request.WithContext(ctx)
					resultHandler = handler
					break
				}
			}
		}
	}
	if resultHandler == nil {
		resultHandler = r.notFoundHandler
	}
	r.mu.RUnlock()

	resultHandler.ServeHTTP(writer, request)
}

func PathParams(ctx context.Context) (*Params, error) {
	params, ok := ctx.Value(paramsContextKey).(*Params)
	if !ok {
		return nil, ErrNoParams
	}
	return params, nil
}

func wrapHandler(handler http.Handler, middlewares... Middleware) http.Handler {
	for _, m := range middlewares {
		handler = m(handler)
	}
	return handler
}

func isValidMethod(method Method) bool {
	for _, m := range []Method{GET, POST, PUT, PATCH, DELETE} {
		if m == method {
			return true
		}
	}
	return false
}
