// Copyright 2013 Ryan Rogers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package routing implements a HTTP request router.
package routing

import (
	"fmt"
	"net/http"
	"strings"
)

// A Router holds all the defined routes, as well as defaults to be used for
// each newly created route.
type Router struct {
	routes          []*Route
	namedRoutes     map[*Route]string
	notFoundHandler http.HandlerFunc
	schemes         map[string]bool // Default schemes applied to all routes
	host            *hostInfo       // Default host name applied to all routes
	matchSlashes    bool
	err             error
}

// A Request contains information relating to the currently matched HTTP
// request.
type Request struct {
	Request *http.Request
	Route   *Route
	Params  map[string]string
}

// NewRouter returns a new Router.
func NewRouter() *Router {
	router := &Router{
		namedRoutes: make(map[*Route]string),
	}
	return router
}

// SetNotFound sets the handler to be used when no routes match a request.
func (r *Router) SetNotFound(f http.HandlerFunc) *Router {
	r.notFoundHandler = f
	return r
}

// NotFound returns the handler used when no routes match a request.
func (r *Router) NotFound() http.HandlerFunc {
	return r.notFoundHandler
}

// SetHost sets a host name that will be applied to all newly created routes.
func (r *Router) SetHost(h string) *Router {
	host, err := parseHost(h)
	if err != nil {
		r.err = err
		return r
	}
	r.host = host
	return r
}

// Host returns the host name that will be applied to all newly created routes.
func (r *Router) Host() string {
	if r.host == nil {
		return ""
	}
	return r.host.rawHost
}

// UnsetHost clears the host name that will be applied to all newly created
// routes.
func (r *Router) UnsetHost() {
	r.host = nil
}

// SetSchemes sets a list of schemes that will be applied to all newly created
// routes.  If an unsupported scheme is provided, no schemes are set, and an
// error message is set on the router.
func (r *Router) SetSchemes(s ...string) *Router {
	schemes, err := validateSchemes(s...)
	if err != nil {
		r.err = err
		return r
	}
	r.schemes = schemes
	return r
}

// Schemes returns the list of schemes that will be applied to all newly
// created routes.
func (r *Router) Schemes() []string {
	s := make([]string, 0, len(r.schemes))
	for k := range r.schemes {
		s = append(s, k)
	}
	return s
}

// UnsetSchemes clears the list of schemes that will be applied to all newly
// created routes.
func (r *Router) UnsetSchemes() {
	r.schemes = nil
}

// SetMatchSlashes sets the default handling of trailing slashes on paths.
// If matchSlashes is true, the request handler will redirect a matched route
// based on whether or not it should have a trailing slash.  For example:
//
// If matchSlashes is true, and a route has a path of "/blog/", a request for
// "/blog" will match, and be redirected to "/blog/".
// Likewise, if the route has a path of "/blog", a request for "/blog/" will
// match, and be redirected to "/blog".
//
// This has no effect if the path is "/".
func (r *Router) SetMatchSlashes(b bool) *Router {
	r.matchSlashes = b
	return r
}

// MatchSlashes returns the status of matchSlashes.
func (r *Router) MatchSlashes() bool {
	return r.matchSlashes
}

// NewRoute creates a new Route using defaults supplied by SetSchemes(),
// SetHost(), and SetMatchSlashes().
func (r *Router) NewRoute() *Route {
	route := &Route{
		router:       r,
		schemes:      r.schemes,
		host:         r.host,
		matchSlashes: r.matchSlashes,
	}
	r.routes = append(r.routes, route)
	return route
}

// Route returns the route named by n.  If no route with that name exists, an
// error is returned.
func (r *Router) Route(n string) (*Route, error) {
	for route, name := range r.namedRoutes {
		if n == name {
			return route, nil
		}
	}
	return nil, fmt.Errorf(errRouteNotDefined, n)
}

// Error returns the last router error that occurred.
func (r *Router) Error() error {
	return r.err
}

// UnsetError removes any error set on the router.
func (r *Router) UnsetError() {
	r.err = nil
}

// ServeHTTP accepts incoming requests and attempts to find a route that
// matches it.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Clean up the Request path.
	// Borrowed from net/http/server.go
	if req.Method != "CONNECT" {
		// Clean path to canonical form and redirect.
		if p := cleanPath(req.URL.Path); p != req.URL.Path {
			w.Header().Set("Location", p)
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}
	}

	r.handleRequest(w, req, r.routes)
}

// handleRequest attempts to find a route that matches the current request,
// then takes the proper steps to send the request to the route's handler.
func (r *Router) handleRequest(w http.ResponseWriter, req *http.Request, routes []*Route) {
	// See if there are any routes that match the request.
	route := match(req, routes)
	if route == nil {
		if r.notFoundHandler == nil {
			http.NotFound(w, req)
		} else {
			r.notFoundHandler(w, req)
		}
		return
	}

	// Redirect to clean up trailing slashes if needed.
	if route.path != nil && route.matchSlashes {
		if strings.HasSuffix(route.path.rawPath, "/") && !strings.HasSuffix(req.URL.Path, "/") {
			http.Redirect(w, req, req.URL.Path+"/", http.StatusMovedPermanently)
			return
		} else if !strings.HasSuffix(route.path.rawPath, "/") && strings.HasSuffix(req.URL.Path, "/") {
			http.Redirect(w, req, req.URL.Path[:len(req.URL.Path)-1], http.StatusMovedPermanently)
			return
		}
	}

	// If the route has a handler defined, call it.
	if route.handler != nil {
		params, err := route.getPathParams(req.URL.Path)
		if err != nil {
			// FIXME: Is a panic the best way to handle an error here?
			panic(err)
		}
		route.handler(w, &Request{
			Request: req,
			Route:   route,
			Params:  params,
		})
	}

	// Handle any child routes.
	if len(route.children) > 0 {
		r.handleRequest(w, req, route.children)
	}
}
