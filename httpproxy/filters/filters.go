package filters

import (
	"fmt"
	"net/http"
)

type Filter interface {
	FilterName() string
}

type RequestFilter interface {
	FilterName() string
	Request(*Context, *http.Request) (*Context, *http.Request, error)
}

type RoundTripFilter interface {
	FilterName() string
	RoundTrip(*Context, *http.Request) (*Context, *http.Response, error)
}

type ResponseFilter interface {
	FilterName() string
	Response(*Context, *http.Response) (*Context, *http.Response, error)
}

type RequestRoundTripFilter interface {
	FilterName() string
	Request(*Context, *http.Request) (*Context, *http.Request, error)
	RoundTrip(*Context, *http.Request) (*Context, *http.Response, error)
}

type RoundTripResponseFilter interface {
	FilterName() string
	RoundTrip(*Context, *http.Request) (*Context, *http.Response, error)
	Response(*Context, *http.Response) (*Context, *http.Response, error)
}

type RequestResponseFilter interface {
	FilterName() string
	Request(*Context, *http.Request) (*Context, *http.Request, error)
	Response(*Context, *http.Response) (*Context, *http.Response, error)
}

type RequestRoundTripResponseFilter interface {
	FilterName() string
	Request(*Context, *http.Request) (*Context, *http.Request, error)
	RoundTrip(*Context, *http.Request) (*Context, *http.Response, error)
	Response(*Context, *http.Response) (*Context, *http.Response, error)
}

type RegisteredFilter struct {
	New func() (Filter, error)
}

var (
	filters map[string]*RegisteredFilter
)

func init() {
	filters = make(map[string]*RegisteredFilter)
}

// Register a Filter
func Register(name string, registeredFilter *RegisteredFilter) error {
	if _, exists := filters[name]; exists {
		return fmt.Errorf("Name already registered %s", name)
	}

	filters[name] = registeredFilter
	return nil
}

// NewFilter creates a new Filter of type "name"
func NewFilter(name string) (Filter, error) {
	filter, exists := filters[name]
	if !exists {
		return nil, fmt.Errorf("filters: Unknown filter %q", name)
	}
	return filter.New()
}
