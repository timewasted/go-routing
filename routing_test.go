// Copyright 2013 Ryan Rogers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
)

//
// Router/Route shared tests
//

func TestHost_invalid(t *testing.T) {
	hosts := []string{
		// Empty hosts are not valid.
		"",
		// More opening braces than closing braces.
		"{{[a-z]+}.example.com",
		// More closing braces than opening braces.
		"{[a-z]+}.example.com}",
		// Regular expression doesn't compile due to missing closing ')'.
		"example.{([a-z]+}",
	}
	router := NewRouter()
	var route *Route

	for _, h := range hosts {
		router.UnsetError()
		router.SetHost(h)
		if router.Error() == nil {
			t.Errorf("%v: Expected an error, received none.", h)
		}
		if router.Host() != "" {
			t.Errorf("%v: Expected empty host, received '%v'.", h, router.Host())
		}

		// The Route should not inherit the host from the Router.
		route = router.NewRoute()
		if route.Host() != "" {
			t.Errorf("%v: Expected empty host, received '%v'.", h, route.Host())
		}

		// Setting the Route's host directly should also fail.
		route.SetHost(h)
		if route.Error() == nil {
			t.Errorf("%v: Expected an error, received none.", h)
		}
		if route.Host() != "" {
			t.Errorf("%v: Expected empty host, received '%v'.", h, route.Host())
		}
	}
}

func TestHost_valid(t *testing.T) {
	hosts := []string{
		"example.com",
		"{[a-z]+}.example.com",
	}
	router := NewRouter()
	var route *Route

	for _, h := range hosts {
		router.UnsetError()
		router.SetHost(h)
		if router.Error() != nil {
			t.Errorf("%v: Expected no error, received '%v'.", h, router.Error())
		}
		if router.Host() != h {
			t.Errorf("Expected host '%v', received '%v'.", h, router.Host())
		}

		// The Route should inherit the host from the Router.
		route = router.NewRoute()
		if route.Host() != h {
			t.Errorf("Expected host '%v', received '%v'.", h, route.Host())
		}

		// Setting the Route's host directly should also succeed.
		route.SetHost(h)
		if route.Error() != nil {
			t.Errorf("%v: Expected no error, received '%v'.", h, route.Error())
		}
		if route.Host() != h {
			t.Errorf("Expected host '%v', received '%v'.", h, route.Host())
		}
	}
}

func TestHost(t *testing.T) {
	router := NewRouter()
	route := router.NewRoute()

	// The default state is empty.
	if router.Host() != "" {
		t.Errorf("Expected empty host, received '%v'.", router.Host())
	}
	if route.Host() != "" {
		t.Errorf("Expected empty host, received '%v'.", route.Host())
	}

	// Host can be unset.
	router.SetHost("example.com")
	router.UnsetHost()
	if router.Host() != "" {
		t.Errorf("Expected empty host, received '%v'.", router.Host())
	}

	route.SetHost("example.com")
	route.UnsetHost()
	if route.Host() != "" {
		t.Errorf("Expected empty host, received '%v'.", route.Host())
	}
}

func TestSchemes_invalid(t *testing.T) {
	schemes := [][]string{
		{""},
		{"ftp"},
		{"gopher"},
	}
	router := NewRouter()
	var route *Route

	for _, s := range schemes {
		router.UnsetError()
		router.SetSchemes(s...)
		if router.Error() == nil {
			t.Errorf("%v: Expected an error, received none.", s)
		}
		if len(router.Schemes()) != 0 {
			t.Errorf("%v: Expected empty schemes, received '%v'.", s, router.Schemes())
		}

		// The Route should not inherit the schemes from the Router.
		route = router.NewRoute()
		if len(route.Schemes()) != 0 {
			t.Errorf("%v: Expected empty schemes, received '%v'.", s, route.Schemes())
		}

		// Setting the Route's schemes directly should also fail.
		route.SetSchemes(s...)
		if route.Error() == nil {
			t.Errorf("%v: Expected an error, received none.", s)
		}
		if len(route.Schemes()) != 0 {
			t.Errorf("%v: Expected empty schemes, received '%v'.", s, route.Schemes())
		}
	}
}

func TestSchemes_valid(t *testing.T) {
	schemes := [][]string{
		{"http"},
		{"https"},
		{"http", "https"},
		{"https", "http"},
	}
	router := NewRouter()
	var route *Route

	for _, s := range schemes {
		router.UnsetError()
		router.SetSchemes(s...)
		if router.Error() != nil {
			t.Errorf("%v: Expected no error, received '%v'.", s, router.Error())
		}
		if len(router.Schemes()) == 0 {
			t.Errorf("Expected schemes '%v', received none.", s)
		}
		if !slicesAreSimilar(s, router.Schemes()) {
			t.Errorf("Expected schemes '%v' to be similar to '%v'.", router.Schemes(), s)
		}

		// The Route should inherit the schemes from the Router.
		route = router.NewRoute()
		if len(route.Schemes()) == 0 {
			t.Errorf("Expected schemes '%v', received none.", s)
		}
		if !slicesAreSimilar(s, route.Schemes()) {
			t.Errorf("Expected schemes '%v' to be similar to '%v'.", route.Schemes(), s)
		}

		// Setting the Route's schemes directly should also succeed.
		route.SetSchemes(s...)
		if route.Error() != nil {
			t.Errorf("%v: Expected no error, received '%v'.", s, route.Error())
		}
		if len(route.Schemes()) == 0 {
			t.Errorf("Expected schemes '%v', received none.", s)
		}
		if !slicesAreSimilar(s, route.Schemes()) {
			t.Errorf("Expected schemes '%v' to be similar to '%v'.", route.Schemes(), s)
		}
	}
}

func TestSchemes(t *testing.T) {
	router := NewRouter()
	route := router.NewRoute()

	// The default state is empty.
	if len(router.Schemes()) != 0 {
		t.Errorf("Expected empty schemes, received '%v'.", router.Schemes())
	}
	if len(route.Schemes()) != 0 {
		t.Errorf("Expected empty schemes, received '%v'.", route.Schemes())
	}

	// The schemes are not case sensitive.
	schemes := []string{"http", "https"}
	router.SetSchemes("hTtP", "HtTpS")
	if router.Error() != nil {
		t.Errorf("Expected no error, received '%v'.", router.Error())
	}
	if len(router.Schemes()) == 0 {
		t.Errorf("Expected schemes '%v', received none.", schemes)
	}
	if !slicesAreSimilar(schemes, router.Schemes()) {
		t.Errorf("Expected schemes '%v' to be similar to '%v'.", router.Schemes(), schemes)
	}

	route.SetSchemes("hTtP", "HtTpS")
	if route.Error() != nil {
		t.Errorf("Expected no error, received '%v'.", route.Error())
	}
	if len(route.Schemes()) == 0 {
		t.Errorf("Expected schemes '%v', received none.", schemes)
	}
	if !slicesAreSimilar(schemes, route.Schemes()) {
		t.Errorf("Expected schemes '%v' to be similar to '%v'.", route.Schemes(), schemes)
	}

	// Must specify at least one scheme.
	router.SetSchemes()
	if router.Error() == nil {
		t.Error("Expected an error, received none.")
	}
	route.SetSchemes()
	if route.Error() == nil {
		t.Error("Expected an error, received none.")
	}

	// Schemes can be unset.
	router.SetSchemes("http")
	router.UnsetSchemes()
	if len(router.Schemes()) != 0 {
		t.Errorf("Expected no schemes, received '%v'.", router.Schemes())
	}

	route.SetSchemes("http")
	route.UnsetSchemes()
	if len(route.Schemes()) != 0 {
		t.Errorf("Expected no schemes, received '%v'.", route.Schemes())
	}
}

func TestMatchSlashes(t *testing.T) {
	router := NewRouter()
	route := router.NewRoute()

	// The default state is false.
	if router.MatchSlashes() {
		t.Error("Expected MatchSlashes to be false, received true.")
	}
	if route.MatchSlashes() {
		t.Error("Expected MatchSlashes to be false, received true.")
	}

	router.SetMatchSlashes(true)
	if !router.MatchSlashes() {
		t.Error("Expected MatchSlashes to be true, received false.")
	}
	route.SetMatchSlashes(true)
	if !route.MatchSlashes() {
		t.Error("Expected MatchSlashes to be true, received false.")
	}
}

func TestUnsetError(t *testing.T) {
	router := NewRouter()
	route := router.NewRoute()

	// The default state is nil.
	if router.Error() != nil {
		t.Errorf("Expected no error, received '%v'.", router.Error())
	}
	if route.Error() != nil {
		t.Errorf("Expected no error, received '%v'.", route.Error())
	}

	// Artificially set an error, then clear it.
	router.err = fmt.Errorf("Test error")
	if router.Error() == nil {
		t.Error("Expected an error, received none.")
	}
	router.UnsetError()
	if router.Error() != nil {
		t.Errorf("Expected no error, received '%v'.", router.Error())
	}

	route.err = fmt.Errorf("Test error")
	if route.Error() == nil {
		t.Error("Expected an error, received none.")
	}
	route.UnsetError()
	if route.Error() != nil {
		t.Errorf("Expected no error, received '%v'.", route.Error())
	}
}

//
// Router tests
//

func TestRouterRoute(t *testing.T) {
	router := NewRouter()
	var route *Route
	var err error

	// No route named "test" exists
	route, err = router.Route("test")
	if err == nil {
		t.Error("Expected an error, received none.")
	}

	// Create route "test"
	testRoute := router.NewRoute().SetName("test")
	route, err = router.Route("test")
	if err != nil {
		t.Errorf("Expected no error, received '%v'.", err)
	}
	if route != testRoute {
		t.Errorf("Expected route '%v', received '%v'.", testRoute, route)
	}
}

func TestRouterHandleRequest(t *testing.T) {
	// FIXME: I think this should probably be tested in some way, but I'm not
	// entirely sure how to test it, or even what needs to be tested.
}

//
// Route tests
//

func TestRouteName(t *testing.T) {
	router := NewRouter()
	route1 := router.NewRoute()
	route2 := router.NewRoute()
	name := "test"

	route1.SetName(name)
	if route1.Error() != nil {
		t.Fatalf("Expected no error, received '%v'.", route1.Error())
	}
	if route1.Name() != name {
		t.Fatalf("Expected name '%v', received '%v'.", name, route1.Name())
	}

	// Route names must be unique across a router.
	route2.SetName(name)
	if route2.Error() == nil {
		t.Error("Expected an error, received none.")
	}
	if route2.Name() != "" {
		t.Errorf("Expected empty name, received '%v'.", route2.Name())
	}

	// Names can be unset.
	route1.UnsetName()
	if route1.Name() != "" {
		t.Errorf("Expected empty name, received '%v'.", route1.Name())
	}
}

func TestRouteMethods_invalid(t *testing.T) {
	methods := [][]string{
		{""},
		{"CERTAINLY"},
		{"NOT"},
		{"VALID"},
	}
	router := NewRouter()
	route := router.NewRoute()

	for _, m := range methods {
		route.UnsetError()
		route.SetMethods(m...)
		if route.Error() == nil {
			t.Errorf("%v: Expected an error, received none.", m)
		}
		if len(route.Methods()) != 0 {
			t.Errorf("%v: Expected empty methods, received '%v'.", m, route.Methods())
		}
	}
}

func TestRouteMethods_valid(t *testing.T) {
	methods := [][]string{
		{"OPTIONS"},
		{"GET"},
		{"HEAD"},
		{"POST"},
		{"PUT"},
		{"DELETE"},
		{"TRACE"},
		{"CONNECT"},
		{"PATCH"},
		{"GET", "POST", "PUT"},
		{"HEAD", "DELETE", "PATCH"},
	}
	router := NewRouter()
	route := router.NewRoute()

	for _, m := range methods {
		route.UnsetError()
		route.SetMethods(m...)
		if route.Error() != nil {
			t.Errorf("%v: Expected no error, received '%v'.", m, route.Error())
		}
		if len(route.Methods()) == 0 {
			t.Errorf("Expected methods '%v', received none.", m)
		}
		if !slicesAreSimilar(m, route.Methods()) {
			t.Errorf("Expected methods '%v' to be similar to '%v'.", route.Methods(), m)
		}
	}
}

func TestRouteMethods(t *testing.T) {
	router := NewRouter()
	route := router.NewRoute()

	// The default state is empty.
	if len(route.Methods()) != 0 {
		t.Errorf("Expected empty methods, received '%v'.", route.Methods())
	}

	// The methods are not case sensitive.
	methods := []string{"GET", "POST"}
	route.SetMethods("GeT", "pOsT")
	if route.Error() != nil {
		t.Errorf("Expected no error, received '%v'.", route.Error())
	}
	if len(route.Methods()) == 0 {
		t.Errorf("Expected methods '%v', received none.", methods)
	}
	if !slicesAreSimilar(methods, route.Methods()) {
		t.Errorf("Expected methods '%v' to be similar to '%v'.", route.Methods(), methods)
	}

	// Must specify at least one method.
	route.SetMethods()
	if route.Error() == nil {
		t.Error("Expected an error, received none.")
	}

	// Methods can be unset.
	route.SetMethods("GET")
	route.UnsetMethods()
	if len(route.Methods()) != 0 {
		t.Errorf("Expected no methods, received '%v'.", route.Methods())
	}
}

func TestRoutePathPrefix_invalid(t *testing.T) {
	paths := []string{
		// Empty paths are not valid.
		"",
		// Empty name.
		"/{:[a-z]+}/",
		// Parameter name redeclared.
		"/{path:[a-z]+}/{path:[0-9]+}/",
		// More opening braces than closing braces.
		"/{{path:[a-z]+}/",
		// More closing braces than opening braces.
		"/{path:[a-z]+}}/",
		// Regular expression doesn't compile due to missing closing ')'.
		"/{path:([a-z]+}/",
	}
	router := NewRouter()
	route := router.NewRoute()

	for _, p := range paths {
		route.UnsetError()
		route.SetPath(p)
		if route.Error() == nil {
			t.Error("Expected an error, received none.")
		}
		if route.Path() != "" {
			t.Errorf("Expected empty path, received '%v'.", route.Path())
		}

		route.UnsetError()
		route.SetPrefix(p)
		if route.Error() == nil {
			t.Error("Expected an error, received none.")
		}
		if route.Path() != "" {
			t.Errorf("Expected emtpy path, received '%v'.", route.Path())
		}
	}
}

func TestRoutePathPrefix_valid(t *testing.T) {
	paths := []string{
		"/",
		"/{path:[a-z]+}/",
		"/blog/{id:[0-9]+}/{slug:[-a-z]+}/",
		"/blog/{id:[0-9]+}/{slug:[-a-z]+}/{extra:}/",
	}
	router := NewRouter()
	route := router.NewRoute()

	for _, p := range paths {
		route.UnsetError()
		route.SetPath(p)
		if route.Error() != nil {
			t.Errorf("Expected no error, received '%v'.", route.Error())
		}
		if route.Path() != p {
			t.Errorf("Expected path '%v', received '%v'.", p, route.Path())
		}

		route.UnsetError()
		route.SetPrefix(p)
		if route.Error() != nil {
			t.Errorf("Expected no error, received '%v'.", route.Error())
		}
		if route.Path() != p {
			t.Errorf("Expected path '%v', received '%v'.", p, route.Path())
		}
	}
}

func TestRoutePathPrefix(t *testing.T) {
	router := NewRouter()
	route := router.NewRoute()

	// The default state is empty.
	if route.Path() != "" {
		t.Errorf("Expected empty path, received '%v'.", route.Path())
	}

	// Path can be unset.
	route.SetPath("/")
	route.UnsetPath()
	if route.Path() != "" {
		t.Errorf("Expected empty path, received '%v'.", route.Path())
	}
}

func TestRouteHeaders(t *testing.T) {
	headers := [][]string{
		{"Accept-Encoding", "gzip"},
		{"Accept-Encoding", "deflate"},
		{"Dnt", "1"},
		{"X-Requested-With", "XMLHttpRequest"},
	}
	router := NewRouter()
	route := router.NewRoute()

	// The default state is empty.
	if len(route.Headers()) != 0 {
		t.Errorf("Expected no headers, received '%v'.", route.Headers())
	}

	for _, h := range headers {
		route.SetHeader(h[0], h[1])
	}
	routeHeaders := route.Headers()
	var exists bool

	for _, h := range headers {
		if _, exists = routeHeaders[h[0]]; !exists {
			t.Errorf("Expected header '%v' to exist in '%v'.", h[0], routeHeaders)
			continue
		}

		exists = false
		for _, v := range routeHeaders[h[0]] {
			if v == h[1] {
				exists = true
				break
			}
		}
		if !exists {
			t.Errorf("Expected header '%v' to exist in '%v'.", h[1], routeHeaders[h[0]])
		}
	}

	// Headers can be unset.
	route.SetHeader("Dnt", "1")
	route.UnsetHeaders()
	if len(route.Headers()) != 0 {
		t.Errorf("Expected no headers, received '%v'.", route.Headers())
	}
}

func TestRouteHandler(t *testing.T) {
	router := NewRouter()
	route := router.NewRoute()

	// The default state is nil.
	if route.Handler() != nil {
		t.Errorf("Expected no handler, received '%v'.", route.Handler())
	}

	// Handler can be unset.
	route.SetHandler(func(w http.ResponseWriter, r *Request) {})
	route.UnsetHandler()
	if route.Handler() != nil {
		t.Errorf("Expected no handler, received '%v'.", route.Handler())
	}
}

func TestRouteSubroute(t *testing.T) {
	parentPath := "/blog/"
	childPath := "/article/{id:[0-9]+}/"
	combinedPath := "/blog/article/{id:[0-9]+}/"
	router := NewRouter()
	parent := router.NewRoute().SetPath(parentPath)

	// Subroutes do not directly inherit their parent's path.
	child := parent.Subroute()
	if child.Path() != "" {
		t.Errorf("Expected an empty path, received '%v'.", child.Path())
	}

	// The parent's path is used when defining the child's path, however.
	if child.parentPath != parentPath {
		t.Errorf("Expected parent path '%v', received '%v'.", parentPath, child.parentPath)
	}
	child.SetPath(childPath)
	if child.Path() != combinedPath {
		t.Errorf("Expected path '%v', received '%v'.", combinedPath, child.Path())
	}
}

//
// Matcher tests
//

func TestRouteMatchSchemes(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Expected no error, received '%v'.", err)
	}
	router := NewRouter()
	route := router.NewRoute()

	// http route, http request
	route.SetSchemes("http")
	if !route.matchSchemes(request) {
		t.Errorf("Expected schemes '%v' to match 'http' request.", route.Schemes())
	}

	// https route, http request
	route.SetSchemes("https")
	if route.matchSchemes(request) {
		t.Errorf("Expected schemes '%v' to not match 'http' request.", route.Schemes())
	}

	// [http, https] route, http request
	route.SetSchemes("http", "https")
	if !route.matchSchemes(request) {
		t.Errorf("Expected schemes '%v' to match 'http' request.", route.Schemes())
	}

	// http route, https request
	request.TLS = new(tls.ConnectionState)
	route.SetSchemes("http")
	if route.matchSchemes(request) {
		t.Errorf("Expected schemes '%v' to not match 'https' request.", route.Schemes())
	}

	// https route, https request
	route.SetSchemes("https")
	if !route.matchSchemes(request) {
		t.Errorf("Expected schemes '%v' to match 'https' request.", route.Schemes())
	}

	// [http, https] route, https request
	route.SetSchemes("http", "https")
	if !route.matchSchemes(request) {
		t.Errorf("Expected schemes '%v' to match 'https' request.", route.Schemes())
	}
}

func TestRouteMatchMethods(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Expected no error, received '%v'.", err)
	}
	router := NewRouter()
	route := router.NewRoute()

	// GET route, GET request
	route.SetMethods("GET")
	if !route.matchMethods(request) {
		t.Errorf("Expected methods '%v' to match '%v' request.", route.Methods(), request.Method)
	}

	// PUT route, GET request
	route.SetMethods("PUT")
	if route.matchMethods(request) {
		t.Errorf("Expected methods '%v' to not match '%v' request.", route.Methods(), request.Method)
	}

	// [GET, PUT] route, GET request
	route.SetMethods("GET", "PUT")
	if !route.matchMethods(request) {
		t.Errorf("Expected methods '%v' to match '%v' request.", route.Methods(), request.Method)
	}
}

func TestRouteMatchHeaders(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Expected no error, received '%v'.", err)
	}
	request.Header.Add("Accept-Encoding", "gzip,deflate")
	request.Header.Add("Cache-Control", "max-age=0")
	request.Header.Add("Cache-Control", "private")
	request.Header.Add("DNT", "1")
	request.Header.Add("X-Requested-With", "XMLHttpRequest")
	router := NewRouter()
	route := router.NewRoute()

	// Header keys are not case sensitive...
	route.headers = make(http.Header)
	route.SetHeader("x-requested-with", "XMLHttpRequest")
	if !route.matchHeaders(request) {
		t.Errorf("Expected headers '%v' to match the request.", route.Headers())
	}

	// ...but the values are.
	route.headers = make(http.Header)
	route.SetHeader("X-Requested-With", "xmlhttprequest")
	if route.matchHeaders(request) {
		t.Errorf("Expected headers '%v' to not match the request.", route.Headers())
	}

	// Matches are done on whole values only.
	route.headers = make(http.Header)
	route.SetHeader("Accept-Encoding", "gzip")
	if route.matchHeaders(request) {
		t.Errorf("Expected headers '%v' to not match the request.", route.Headers())
	}

	// If a request has multiple values for the same key, we can match any of the values.
	route.headers = make(http.Header)
	route.SetHeader("Cache-Control", "max-age=0")
	if !route.matchHeaders(request) {
		t.Errorf("Expected headers '%v' to match the request.", route.Headers())
	}

	// If a route has multiple values for the same key, all values must match the request.
	route.headers = make(http.Header)
	route.SetHeader("Cache-Control", "max-age=0")
	route.SetHeader("Cache-Control", "public")
	if route.matchHeaders(request) {
		t.Errorf("Expected headers '%v' to not match the request.", route.Headers())
	}

	// All route headers must be present in the request.
	route.headers = make(http.Header)
	route.SetHeader("DNT", "1")
	route.SetHeader("Content-Type", "text/plain")
	if route.matchHeaders(request) {
		t.Errorf("Expected headers '%v' to not match the request.", route.Headers())
	}
}

func TestRouteMatchHost_invalid(t *testing.T) {
	hosts := []string{
		// The port number is stripped off the request before matching.
		"www.example.com:8080",
		// A partial match isn't good enough.
		"example.com",
	}
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Expected no error, received '%v'.", err)
	}
	request.Host = "www.example.com:8080"
	router := NewRouter()
	route := router.NewRoute()

	for _, h := range hosts {
		route.SetHost(h)
		if route.matchHost(request) {
			t.Error("Expected host '%v' to not match the request.", h)
		}
	}
}

func TestRouteMatchHost_valid(t *testing.T) {
	hosts := []string{
		"",
		"www.example.com",
		"{[a-z]+}.example.com",
		"www.example.{[a-z]{2,4}}",
	}
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Expected no error, received '%v'.", err)
	}
	request.Host = "www.example.com:8080"
	router := NewRouter()
	route := router.NewRoute()

	for _, h := range hosts {
		route.SetHost(h)
		if !route.matchHost(request) {
			t.Error("Expected host '%v' to match the request.", h)
		}
	}
}

type matchPathTest struct {
	routePath                 string
	matchPrefix, matchSlashes bool
	requestPaths              []string
}

func TestRouteMatchPath_invalid(t *testing.T) {
	paths := []matchPathTest{
		{ // 0
			routePath:    "/blog",
			matchPrefix:  false,
			matchSlashes: false,
			requestPaths: []string{
				"/",                   // 0
				"/blog/",              // 1
				"/blog/article/1234/", // 2
			},
		},
		{ // 1
			routePath:    "/blog/article/{id:[0-9]+}",
			matchPrefix:  true,
			matchSlashes: true,
			requestPaths: []string{
				"/blog/article/",     // 0
				"/blog/article/abcd", // 1
			},
		},
	}
	router := NewRouter()
	route := router.NewRoute()
	var request *http.Request
	var err error

	for pos1, p := range paths {
		route.UnsetError()
		route.SetMatchSlashes(p.matchSlashes)
		route.path = nil
		if p.routePath != "" {
			if p.matchPrefix {
				route.SetPrefix(p.routePath)
			} else {
				route.SetPath(p.routePath)
			}
			if route.Error() != nil {
				t.Errorf("paths[%v]: Expected no error, received '%v'.", pos1, route.Error())
				continue
			}
		}
		for pos2, r := range p.requestPaths {
			request, err = http.NewRequest("GET", r, nil)
			if err != nil {
				t.Errorf("paths[%v][%v]: Expected no error, received '%v'.", pos1, pos2, err)
				continue
			}
			if route.matchPath(request) {
				t.Errorf("paths[%v][%v]: Expected path '%v' to not match '%v'.", pos1, pos2, p.routePath, r)
			}
		}
	}
}

func TestRouteMatchPath_valid(t *testing.T) {
	paths := []matchPathTest{
		{ // 0
			routePath:    "",
			matchPrefix:  false,
			matchSlashes: false,
			requestPaths: []string{
				"/",            // 0
				"/favicon.ico", // 1
			},
		},
		{ // 1
			routePath:    "/blog",
			matchPrefix:  true,
			matchSlashes: true,
			requestPaths: []string{
				"/blog",               // 0
				"/blog/",              // 1
				"/blog/article/1234/", // 2
			},
		},
		{ // 2
			routePath:    "/blog/article/{id:[0-9]+}",
			matchPrefix:  false,
			matchSlashes: false,
			requestPaths: []string{
				"/blog/article/1",    // 0
				"/blog/article/1234", // 1
			},
		},
	}
	router := NewRouter()
	route := router.NewRoute()
	var request *http.Request
	var err error

	for pos1, p := range paths {
		route.SetMatchSlashes(p.matchSlashes)
		route.path = nil
		if p.routePath != "" {
			if p.matchPrefix {
				route.SetPrefix(p.routePath)
			} else {
				route.SetPath(p.routePath)
			}
			if route.Error() != nil {
				t.Errorf("paths[%v]: Expected no error, received '%v'.", pos1, route.Error())
				continue
			}
		}
		for pos2, r := range p.requestPaths {
			request, err = http.NewRequest("GET", r, nil)
			if err != nil {
				t.Errorf("paths[%v][%v]: Expected no error, received '%v'.", pos1, pos2, err)
				continue
			}
			if !route.matchPath(request) {
				t.Errorf("paths[%v][%v]: Expected path '%v' to match '%v'.", pos1, pos2, p.routePath, r)
			}
		}
	}
}

func TestRouteGetPathParams(t *testing.T) {
	type getPathParamsTest struct {
		routePath   string
		requestPath string
		params      map[string]string
	}

	paths := []getPathParamsTest{
		{ // 0
			routePath:   "/",
			requestPath: "/",
			params:      map[string]string{},
		},
		{ // 1
			routePath:   "/{path:[a-z]+}/",
			requestPath: "/index/",
			params: map[string]string{
				"path": "index",
			},
		},
		{ // 2
			routePath:   "/blog/{id:[0-9]+}/{slug:[-a-z]+}/",
			requestPath: "/blog/1234/super-cool-article/",
			params: map[string]string{
				"id":   "1234",
				"slug": "super-cool-article",
			},
		},
		{ // 3
			routePath:   "/blog/{id:[0-9]+}/{slug:[-a-z]+}/{extra:}/",
			requestPath: "/blog/1234/super-cool-article/misc%20stuff/",
			params: map[string]string{
				"id":    "1234",
				"slug":  "super-cool-article",
				"extra": "misc%20stuff",
			},
		},
		{ // 4
			routePath:   "/item/{id:([a-z]+)([0-9]+)}/",
			requestPath: "/item/abcd1234/",
			params: map[string]string{
				"id": "abcd1234",
			},
		},
	}
	router := NewRouter()
	route := router.NewRoute()
	var params map[string]string
	var err error

	for pos, p := range paths {
		route.UnsetError()
		route.SetPath(p.routePath)
		if route.Error() != nil {
			t.Errorf("paths[%v]: Expected no error, received '%v'.", pos, route.Error())
			continue
		}
		params, err = route.getPathParams(p.requestPath)
		if err != nil {
			t.Errorf("paths[%v]: Expected no error, received '%v'.", pos, err)
			continue
		}
		if len(params) != len(p.params) {
			t.Errorf("paths[%v]: Expected %d params, received '%v'.", pos, len(p.params), params)
			continue
		}
		for k, v := range p.params {
			if _, exists := params[k]; !exists {
				t.Errorf("paths[%v]: Expected key '%v' to exist in '%v'.", pos, k, params)
			} else if params[k] != v {
				t.Errorf("paths[%v]: Expected value '%v', received '%v'.", pos, v, params[k])
			}
		}
	}
}

//
// Helpers
//

func TestHelper_match(t *testing.T) {
	// Each of the individual mapping functions is tested already.  This just
	// needs to make sure it picks the expected route.
	type matchTest struct {
		isTLS   bool
		method  string
		path    string
		headers http.Header
		route   *Route
	}

	router := NewRouter()
	// Define the routes, giving them names for better error messages.
	route1 := router.NewRoute().SetName("route1").
		Get("/")
	router.NewRoute().SetName("route2").
		SetSchemes("https").Get("/")
	route3 := router.NewRoute().SetName("route3").
		Delete("/")
	route4 := router.NewRoute().SetName("route4").
		XHR().SetHeader("If-None-Match", "1234abcd")
	// Define the requests.
	requests := []matchTest{
		{ // 0
			isTLS:   false,
			method:  "GET",
			path:    "/",
			headers: http.Header{},
			route:   route1,
		},
		{ // 1
			// This is matching route1 instead of route2 due to matching being
			// done in first defined, first served order.
			isTLS:   true,
			method:  "GET",
			path:    "/",
			headers: http.Header{},
			route:   route1,
		},
		{ // 2
			isTLS:   false,
			method:  "DELETE",
			path:    "/nonexistent/",
			headers: http.Header{},
			route:   nil,
		},
		{ // 3
			isTLS:   true,
			method:  "DELETE",
			path:    "/",
			headers: http.Header{},
			route:   route3,
		},
		{ // 4
			// This doesn't match route4 because it also requires
			// "If-None-Match" to be present and correct.
			isTLS:  false,
			method: "PUT",
			path:   "/nonexistent/",
			headers: http.Header{
				"X-Requested-With": {"XMLHttpRequest"},
			},
			route: nil,
		},
		{ // 4
			isTLS:  false,
			method: "PUT",
			path:   "/nonexistent/",
			headers: http.Header{
				"X-Requested-With": {"XMLHttpRequest"},
				"If-None-Match":    {"1234abcd"},
			},
			route: route4,
		},
	}
	tlsState := new(tls.ConnectionState)
	var matched *Route
	var request *http.Request
	var err error

	for pos, r := range requests {
		request, err = http.NewRequest(r.method, r.path, nil)
		if err != nil {
			t.Errorf("Expected no error, received '%v'.", err)
			continue
		}
		if r.isTLS {
			request.TLS = tlsState
		}
		if len(r.headers) != 0 {
			request.Header = r.headers
		}
		matched = match(request, router.routes)
		if matched != r.route {
			if matched == nil && r.route == nil {
				t.Errorf("requests[%v]: Expected route to match.", pos)
			} else if matched == nil {
				t.Errorf("requests[%v]: Expected route '%v' to match.", pos, r.route.Name())
			} else if r.route == nil {
				t.Errorf("requests[%v]: Expected route to match, received '%v'.", pos, matched.Name())
			} else {
				t.Errorf("requests[%v]: Expected route '%v' to match, received '%v'.", pos, r.route.Name(), matched.Name())
			}
		}
	}
}

func TestHelper_parseHost_invalid(t *testing.T) {
	hosts := []string{
		// Empty hosts are not valid.
		"",
		// More opening braces than closing braces.
		"{{[a-z]+}.example.com",
		// More closing braces than opening braces.
		"{[a-z]+}.example.com}",
		// Regular expression doesn't compile due to missing closing ')'.
		"example.{([a-z]+}",
	}

	for _, h := range hosts {
		if _, err := parseHost(h); err == nil {
			t.Errorf("Expected an error from host '%v', received none.", h)
		}
	}
}

func TestHelper_parseHost_valid(t *testing.T) {
	hosts := [][]string{
		{
			"{[a-z]+}.example.com",
			"^([a-z]+)\\.example\\.com$",
		},
		{
			"{[a-z]+}.example.{(com|org)}",
			"^([a-z]+)\\.example\\.((com|org))$",
		},
		{
			"{([a-z]+)([0-9]+)}.example.{[a-z]{2,4}}",
			"^(([a-z]+)([0-9]+))\\.example\\.([a-z]{2,4})$",
		},
	}
	var parsedHost *hostInfo
	var err error

	for _, h := range hosts {
		if parsedHost, err = parseHost(h[0]); err != nil {
			t.Errorf("Expected no error from host '%v', received '%v'.", h[0], err)
			continue
		}
		if parsedHost.rawHost != h[0] {
			t.Errorf("Expected host '%v', received '%v'.", h[0], parsedHost.rawHost)
		}
		if parsedHost.pattern.String() != h[1] {
			t.Errorf("Expected pattern '%v', received '%v'.", h[1], parsedHost.pattern.String())
		}
	}
}

type parsePathTest struct {
	matchPrefix  bool
	matchSlashes bool
	rawPath      string
	fwdPattern   string
	revPattern   string
	params       [][]string
}

func TestHelper_parsePath_invalid(t *testing.T) {
	paths := []string{
		// Empty paths are not valid.
		"",
		// Empty name.
		"/{:[a-z]+}/",
		// Parameter name redeclared.
		"/{path:[a-z]+}/{path:[0-9]+}/",
		// More opening braces than closing braces.
		"/{{path:[a-z]+}/",
		// More closing braces than opening braces.
		"/{path:[a-z]+}}/",
		// Regular expression doesn't compile due to missing closing ')'.
		"/{path:([a-z]+}/",
	}

	for _, p := range paths {
		// matchPrefix, matchSlashes = false, true (should have no bearing on tests).
		if _, err := parsePath(p, false, true); err == nil {
			t.Errorf("Expected an error from path '%v', received none.", p)
		}
	}
}

func TestHelper_parsePath_valid(t *testing.T) {
	paths := []parsePathTest{
		{ // 0
			matchPrefix:  false,
			matchSlashes: false,
			rawPath:      "/",
			fwdPattern:   "^/$",
			revPattern:   "/",
			params:       [][]string{},
		},
		{ // 1
			matchPrefix:  true,
			matchSlashes: true,
			rawPath:      "/{path:[a-z]+}/",
			fwdPattern:   "^/([a-z]+)/?",
			revPattern:   "/%s/",
			params: [][]string{
				{"path", "[a-z]+"},
			},
		},
		{ // 2
			matchPrefix:  false,
			matchSlashes: false,
			rawPath:      "/{path:[a-z]+}/",
			fwdPattern:   "^/([a-z]+)/$",
			revPattern:   "/%s/",
			params: [][]string{
				{"path", "[a-z]+"},
			},
		},
		{ // 3
			matchPrefix:  true,
			matchSlashes: true,
			rawPath:      "/blog/{id:[0-9]+}/{slug:[-a-z]+}/",
			fwdPattern:   "^/blog/([0-9]+)/([-a-z]+)/?",
			revPattern:   "/blog/%s/%s/",
			params: [][]string{
				{"id", "[0-9]+"},
				{"slug", "[-a-z]+"},
			},
		},
		{ // 4
			matchPrefix:  false,
			matchSlashes: false,
			rawPath:      "/blog/{id:[0-9]+}/{slug:[-a-z]+}/",
			fwdPattern:   "^/blog/([0-9]+)/([-a-z]+)/$",
			revPattern:   "/blog/%s/%s/",
			params: [][]string{
				{"id", "[0-9]+"},
				{"slug", "[-a-z]+"},
			},
		},
		{ // 5
			matchPrefix:  true,
			matchSlashes: true,
			rawPath:      "/blog/{id:[0-9]+}/{slug:[-a-z]+}/{extra:}/",
			fwdPattern:   "^/blog/([0-9]+)/([-a-z]+)/([^/]+)/?",
			revPattern:   "/blog/%s/%s/%s/",
			params: [][]string{
				{"id", "[0-9]+"},
				{"slug", "[-a-z]+"},
				{"extra", "[^/]+"},
			},
		},
		{ // 6
			matchPrefix:  false,
			matchSlashes: false,
			rawPath:      "/blog/{id:[0-9]+}/{slug:[-a-z]+}/{extra:}/",
			fwdPattern:   "^/blog/([0-9]+)/([-a-z]+)/([^/]+)/$",
			revPattern:   "/blog/%s/%s/%s/",
			params: [][]string{
				{"id", "[0-9]+"},
				{"slug", "[-a-z]+"},
				{"extra", "[^/]+"},
			},
		},
	}
	var parsedPath *pathInfo
	var err error

	for pos, p := range paths {
		// Make sure there are no errors
		parsedPath, err = parsePath(p.rawPath, p.matchPrefix, p.matchSlashes)
		if err != nil {
			t.Errorf("paths[%v]: Expected no error, received '%v'.", pos, err)
			continue
		}
		// Ensure that rawPath is correct
		if parsedPath.rawPath != p.rawPath {
			t.Errorf("paths[%v]: Expected rawPath '%v', received '%v'.", pos, p.rawPath, parsedPath.rawPath)
		}
		// Ensure that fwdPattern is correct
		if parsedPath.fwdPattern.String() != p.fwdPattern {
			t.Errorf("paths[%v]: Expected fwdPattern '%v', received '%v'.", pos, p.fwdPattern, parsedPath.fwdPattern.String())
		}
		// Ensure that revPattern is correct
		if parsedPath.revPattern != p.revPattern {
			t.Errorf("paths[%v]: Expected revPattern '%v', received '%v'.", pos, p.revPattern, parsedPath.revPattern)
		}
		// Ensure that params are correct
		if len(parsedPath.params) != len(p.params) {
			t.Errorf("paths[%v]: Expected %d params, received '%v'.", pos, len(p.params), parsedPath.params)
			continue
		}
		for k, v := range p.params {
			if parsedPath.params[k][0] != v[0] {
				t.Errorf("paths[%v]: Expected param name '%v', received '%v'.", pos, v[0], parsedPath.params[k][0])
			}
			if parsedPath.params[k][1] != v[1] {
				t.Errorf("paths[%v]: Expected param value '%v', received '%v'.", pos, v[1], parsedPath.params[k][1])
			}
		}
	}
}

func slicesAreSimilar(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	similar := false
	for _, a := range s1 {
		for _, b := range s2 {
			if a == b {
				similar = true
				break
			}
		}
		if !similar {
			return false
		}
		similar = false
	}
	return true
}
