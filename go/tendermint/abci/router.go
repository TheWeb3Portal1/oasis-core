package abci

import (
	"reflect"
	"strings"

	"github.com/tendermint/tendermint/abci/types"

	"github.com/oasislabs/ekiden/go/common/cbor"
	"github.com/oasislabs/ekiden/go/tendermint/api"
)

// QueryHandler is a query handler function.
type QueryHandler func(state interface{}, request interface{}) ([]byte, error)

// QueryRouter is a query router.
type QueryRouter interface {
	// WithApp configures the query router for the given application.
	WithApp(Application) QueryRouter

	// AddRoute adds a route handler.
	//
	// As reflection is used, requestType should be an instance of the
	// actual type used for the request (as opposed to a pointer).
	AddRoute(path string, requestType interface{}, handler QueryHandler)

	// Route performs routing of a given request.
	Route(request types.RequestQuery) types.ResponseQuery
}

type route struct {
	app         Application
	path        string
	requestType *reflect.Type
	handler     QueryHandler
}

func (r *route) handle(request types.RequestQuery) types.ResponseQuery {
	var rq interface{}

	if r.requestType != nil {
		// Using `reflect` is somewhat frowned upon but whatever
		// overhead there may be (< 1 usec), is dwarfed by the
		// memory allocation(s) and CBOR parsing.
		rq = reflect.New(*r.requestType).Interface()
		if err := cbor.Unmarshal(request.GetData(), rq); err != nil {
			return types.ResponseQuery{
				Code: api.CodeInvalidFormat.ToInt(),
			}
		}
	}

	// Get state snapshot based on specified version.
	state, err := r.app.GetState(request.GetHeight())
	if err != nil {
		return types.ResponseQuery{
			Code: api.CodeTransactionFailed.ToInt(),
			Info: err.Error(),
		}
	}

	response, err := r.handler(state, rq)
	if err != nil {
		return types.ResponseQuery{
			Code: api.CodeTransactionFailed.ToInt(),
			Info: err.Error(),
		}
	}
	if response == nil {
		return types.ResponseQuery{
			Code: api.CodeNotFound.ToInt(),
		}
	}

	return types.ResponseQuery{
		Code:  api.CodeOK.ToInt(),
		Value: response,
	}
}

type queryRouter struct {
	routes []route
}

func (r *queryRouter) WithApp(app Application) QueryRouter {
	return &applicationQueryRouter{
		queryRouter: r,
		app:         app,
	}
}

func (r *queryRouter) AddRoute(path string, requestType interface{}, handler QueryHandler) {
	panic("router: no application configured, call WithApp first")
}

func (r *queryRouter) Route(request types.RequestQuery) types.ResponseQuery {
	for _, route := range r.routes {
		if route.path == request.GetPath() {
			return route.handle(request)
		}
	}

	// No routes matched.
	return types.ResponseQuery{
		Code: api.CodeNotFound.ToInt(),
	}
}

type applicationQueryRouter struct {
	*queryRouter

	app Application
}

func (r *applicationQueryRouter) WithApp(app Application) QueryRouter {
	panic("router: application already configured")
}

func (r *applicationQueryRouter) Route(request types.RequestQuery) types.ResponseQuery {
	for _, route := range r.routes {
		if route.app != r.app {
			continue
		}

		if route.path == request.GetPath() {
			return route.handle(request)
		}
	}

	// No routes matched.
	return types.ResponseQuery{
		Code: api.CodeNotFound.ToInt(),
	}
}

func (r *applicationQueryRouter) AddRoute(path string, requestType interface{}, handler QueryHandler) {
	if !isP2PFilterQuery(path) && !strings.HasPrefix(path, r.app.Name()) {
		panic("router: route must start with application name")
	}

	rt := route{
		app:     r.app,
		path:    path,
		handler: handler,
	}
	if requestType != nil {
		tmp := reflect.TypeOf(requestType)
		rt.requestType = &tmp
	}
	r.routes = append(r.routes, rt)
}

// NewQueryRouter constructs a new query router.
func NewQueryRouter() QueryRouter {
	return &queryRouter{}
}
