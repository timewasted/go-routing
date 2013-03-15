// Copyright 2013 Ryan Rogers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// A HandlerFunc is the function signature of the handler that is called when
// a route matches a request.
type HandlerFunc func(http.ResponseWriter, *Request)

// A Route holds all the information about a route.
type Route struct {
	router       *Router
	schemes      map[string]bool
	host         *hostInfo
	methods      map[string]bool
	parentPath   string
	path         *pathInfo
	headers      http.Header
	matchSlashes bool
	handler      HandlerFunc
	children     []*Route
	err          error
}

// SetName sets a name for the route.  Route names must be unique across the
// router.  If the name is already in use, an error is set on the route.
func (r *Route) SetName(n string) *Route {
	for _, v := range r.router.namedRoutes {
		if n == v {
			r.err = fmt.Errorf(errRouteAlreadyDefined, n)
			return r
		}
	}
	r.router.namedRoutes[r] = n
	return r
}

// Name returns the name of the route.  If the route is not named, an empty
// string is returned.
func (r *Route) Name() string {
	return r.router.namedRoutes[r]
}

// UnsetName clears the name assigned to this route.
func (r *Route) UnsetName() {
	delete(r.router.namedRoutes, r)
}

// SetSchemes sets a list of schemes that the route will match.  At least one
// of the provided schemes must match for the route to match a request.  If an
// unsupported scheme is provided, no schemes are set, and an error message
// is set on the route.
func (r *Route) SetSchemes(s ...string) *Route {
	schemes, err := validateSchemes(s...)
	if err != nil {
		r.err = err
		return r
	}
	r.schemes = schemes
	return r
}

// Schemes returns the list of schemes that the route will match.
func (r *Route) Schemes() []string {
	s := make([]string, 0, len(r.schemes))
	for k := range r.schemes {
		s = append(s, k)
	}
	return s
}

// UnsetSchemes clears the list of schemes that the route will match.
func (r *Route) UnsetSchemes() {
	r.schemes = nil
}

// SetHost sets the host name that the route will match.
func (r *Route) SetHost(h string) *Route {
	host, err := parseHost(h)
	if err != nil {
		r.err = err
		return r
	}
	r.host = host
	return r
}

// Host returns the host name that the route will match.
func (r *Route) Host() string {
	if r.host == nil {
		return ""
	}
	return r.host.rawHost
}

// UnsetHost clears the host name that the route will match.
func (r *Route) UnsetHost() {
	r.host = nil
}

// SetMethods sets a list of methods that the route will match.  At least one
// of the provided methods must match for the route to match a request.  If an
// unsupported method is provided, no methods are set, and an error message
// is set on the route.
func (r *Route) SetMethods(m ...string) *Route {
	methods, err := validateMethods(m...)
	if err != nil {
		r.err = err
		return r
	}
	r.methods = methods
	return r
}

// Methods returns the list of methods that the route will match.
func (r *Route) Methods() []string {
	m := make([]string, 0, len(r.methods))
	for k := range r.methods {
		m = append(m, k)
	}
	return m
}

// UnsetMethods clears the list of methods that the route will match.
func (r *Route) UnsetMethods() {
	r.methods = nil
}

// SetPath sets the path that the route will match.  If parsing of the path
// fails, no path is set, and an error message is set on the route.
func (r *Route) SetPath(p string) *Route {
	return r.setPath(p, false)
}

// SetPrefix sets the path prefix that the route will match.  If parsing of
// the path fails, no path is set, and an error message is set on the route.
func (r *Route) SetPrefix(p string) *Route {
	return r.setPath(p, true)
}

// setPath does basic sanity checking of the path, and if the path appears to
// be valid and parses correctly, sets the path of the route.
func (r *Route) setPath(p string, matchPrefix bool) *Route {
	if r.parentPath != "" {
		if strings.HasSuffix(r.parentPath, "/") && strings.HasPrefix(p, "/") {
			p = p[1:]
		}
		p = r.parentPath + p
	}
	parsedPath, err := parsePath(p, matchPrefix, r.matchSlashes)
	if err != nil {
		r.err = err
		return r
	}
	r.path = parsedPath
	return r
}

// Path returns the path that the route will match.
func (r *Route) Path() string {
	if r.path == nil {
		return ""
	}
	return r.path.rawPath
}

// UnsetPath clears the path that the route will match.
func (r *Route) UnsetPath() {
	r.path = nil
}

// SetHeader sets a header name:value pair that the route will match.  Names
// are not case sensitive, but values are.  A name can have multiple values,
// in which case all values are required for the route to match.  Values are
// exact matches.  This means that neither
//
// SetHeader("Accept-Encoding", "gzip")
//
// or
//
// SetHeader("Accept-Encoding", "gzip")
// SetHeader("Accept-Encoding", "deflate")
//
// will match the header "Accept-Encoding: gzip,deflate".  In order to match
// that, you would need to call SetHeader("Accept-Encoding", "gzip,deflate").
func (r *Route) SetHeader(k, v string) *Route {
	if r.headers == nil {
		r.headers = make(http.Header)
	}
	r.headers.Add(k, v)
	return r
}

// Headers returns the list of headers that the route will match.
func (r *Route) Headers() http.Header {
	if r.headers == nil {
		r.headers = make(http.Header)
	}
	return r.headers
}

// UnsetHeaders clears the list of headers that the route will match.
func (r *Route) UnsetHeaders() {
	r.headers = nil
}

// SetMatchSlashes sets the handling of trailing slashes on paths.  See
// Router.SetMatchSlashes for a description of how this works.
func (r *Route) SetMatchSlashes(b bool) *Route {
	r.matchSlashes = b
	return r
}

// MatchSlashes returns the status of matchSlashes.
func (r *Route) MatchSlashes() bool {
	return r.matchSlashes
}

// SetHandler sets the handler that is called when a route is matched.
func (r *Route) SetHandler(f HandlerFunc) *Route {
	r.handler = f
	return r
}

// Handler returns the handler function called when a route matches.
func (r *Route) Handler() HandlerFunc {
	return r.handler
}

// UnsetHandler clears the handler function that is called when a route
// matches.
func (r *Route) UnsetHandler() {
	r.handler = nil
}

// Subroute creates a child Route.
func (r *Route) Subroute() *Route {
	child := r.router.NewRoute()
	r.children = append(r.children, child)
	child.parentPath = r.path.rawPath
	return child
}

// Error returns the last route error that occurred.
func (r *Route) Error() error {
	return r.err
}

// UnsetError removes any error set on the route.
func (r *Route) UnsetError() {
	r.err = nil
}

//
// Shorthand functions
//

// Get is shorthand for SetPath(p) and SetMethods("GET").
func (r *Route) Get(p string) *Route {
	return r.SetPath(p).SetMethods("GET")
}

// GetPrefix is shorthand for SetPrefix(p) and SetMethods("GET").
func (r *Route) GetPrefix(p string) *Route {
	return r.SetPrefix(p).SetMethods("GET")
}

// Head is shorthand for SetPath(p) and SetMethods("HEAD").
func (r *Route) Head(p string) *Route {
	return r.SetPath(p).SetMethods("HEAD")
}

// HeadPrefix is shorthand for SetPrefix(p) and SetMethods("HEAD").
func (r *Route) HeadPrefix(p string) *Route {
	return r.SetPrefix(p).SetMethods("HEAD")
}

// Post is shorthand for SetPath(p) and SetMethods("POST").
func (r *Route) Post(p string) *Route {
	return r.SetPath(p).SetMethods("POST")
}

// PostPrefix is shorthand for SetPrefix(p) and SetMethods("POST").
func (r *Route) PostPrefix(p string) *Route {
	return r.SetPrefix(p).SetMethods("POST")
}

// Put is shorthand for SetPath(p) and SetMethods("PUT").
func (r *Route) Put(p string) *Route {
	return r.SetPath(p).SetMethods("PUT")
}

// PutPrefix is shorthand for SetPrefix(p) and SetMethods("PUT").
func (r *Route) PutPrefix(p string) *Route {
	return r.SetPrefix(p).SetMethods("PUT")
}

// Delete is shorthand for SetPath(p) and SetMethods("DELETE").
func (r *Route) Delete(p string) *Route {
	return r.SetPath(p).SetMethods("DELETE")
}

// DeletePrefix is shorthand for SetPrefix(p) and SetMethods("DELETE").
func (r *Route) DeletePrefix(p string) *Route {
	return r.SetPrefix(p).SetMethods("DELETE")
}

// Patch is shorthand for SetPath(p) and SetMethods("PATCH").
func (r *Route) Patch(p string) *Route {
	return r.SetPath(p).SetMethods("PATCH")
}

// PatchPrefix is shorthand for SetPrefix(p) and SetMethods("PATCH").
func (r *Route) PatchPrefix(p string) *Route {
	return r.SetPrefix(p).SetMethods("PATCH")
}

// XHR is shorthand for SetHeader("X-Requested-With", "XMLHttpRequest").
func (r *Route) XHR() *Route {
	return r.SetHeader("X-Requested-With", "XMLHttpRequest")
}

//
// Matchers
//

// matchSchemes returns true if the route matches the request.
func (r *Route) matchSchemes(req *http.Request) bool {
	if len(r.schemes) > 0 {
		isTLS := req.TLS != nil
		if (isTLS && !r.schemes["https"]) || (!isTLS && !r.schemes["http"]) {
			return false
		}
	}
	return true
}

// matchMethods returns true if the route matches the request.
func (r *Route) matchMethods(req *http.Request) bool {
	if len(r.methods) > 0 && !r.methods[strings.ToUpper(req.Method)] {
		return false
	}
	return true
}

// matchHeaders returns true if the route matches the request.
func (r *Route) matchHeaders(req *http.Request) bool {
	matched := true
	if len(r.headers) > 0 {
		for k, v := range r.headers {
			if _, ok := req.Header[k]; !ok || !sliceContainsStrings(req.Header[k], v) {
				matched = false
				break
			}
		}
	}
	return matched
}

// hostPortRegexp is used to strip the port number off of http.Request.Host.
// FIXME: If the host is an IPv6 address, this will mangle it.
var hostPortRegexp = regexp.MustCompile(":\\d{1,5}$")

// matchHost returns true if the route matches the request.
func (r *Route) matchHost(req *http.Request) bool {
	if r.host != nil {
		host := req.Host
		if hostPortRegexp.MatchString(host) {
			host = host[:strings.LastIndex(host, ":")]
		}
		if !r.host.pattern.MatchString(host) {
			return false
		}
	}
	return true
}

// matchPath returns true if the route matches the request.
func (r *Route) matchPath(req *http.Request) bool {
	if r.path != nil && !r.path.fwdPattern.MatchString(req.URL.Path) {
		return false
	}
	return true
}

//
// Helpers
//

// getPathParams extracts the path parameters from the provided path.
func (r *Route) getPathParams(path string) (map[string]string, error) {
	params := make(map[string]string)
	if r.path.fwdPattern == nil {
		return params, nil
	}
	paramIndex := r.path.fwdPattern.FindStringSubmatchIndex(path)
	if paramIndex == nil {
		return params, nil
	}

	// paramIndex[i] is where the param starts, and paramIndex[i+1] is where it ends.
	// Skip the first pair, since that is just [startOfString, endOfString].
	for i, j := 2, 2; i < len(paramIndex); i += j - i {
		for j += 2; j < len(paramIndex); j += 2 {
			// Skip over all parameters that start and end between [i] and [i+1].
			if paramIndex[i+1] <= paramIndex[j] {
				break
			}
		}
		params[r.path.params[len(params)][0]] = path[paramIndex[i]:paramIndex[i+1]]
	}
	if len(params) != len(r.path.params) {
		return nil, fmt.Errorf(errUnexpectedParamCount, len(r.path.params), len(params))
	}
	return params, nil
}
