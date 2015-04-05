package filters

import (
	"fmt"
	"net/http"
)

type Filter interface {
	FilterName() string
}

type RequestFilter interface {
	Filter
	Request(*Context, *http.Request) (*Context, *http.Request, error)
}

type FetchFilter interface {
	Filter
	Fetch(*Context, *http.Request) (*Context, *http.Response, error)
}

type ResponseFilter interface {
	Filter
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
