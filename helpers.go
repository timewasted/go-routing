// Copyright 2013 Ryan Rogers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
)

// Error messages related to routing.
const (
	errUnsupportedScheme    = "routing: '%s' is not a supported scheme."
	errUnsupportedMethod    = "routing: '%s' is not a supported method."
	errRouteAlreadyDefined  = "routing: Route '%s' is already defined."
	errRouteNotDefined      = "routing: Route '%s' is not defined."
	errPathIsInvalid        = "routing: '%s' is not a valid path."
	errUnexpectedParamCount = "routing: Expected %d params, received %d."
)

// Error messages related to host and path parsing.
const (
	errEmptyHost           = "routing: Host can not be empty."
	errEmptyPath           = "routing: Path can not be empty."
	errUnevenBraces        = "routing: Uneven number of braces."
	errParamNameDefined    = "routing: Parameter '%s' has already been defined."
	errParamNameNotDefined = "routing: Parameter name can not be empty."
)

// hostInfo holds all of the components of a valid parsed host.
type hostInfo struct {
	rawHost string
	pattern *regexp.Regexp
}

// pathInfo holds all of the components of a valid parsed path.
type pathInfo struct {
	rawPath    string
	fwdPattern *regexp.Regexp
	revPattern string // FIXME: This isn't actually used yet.
	params     [][]string
}

// The list of valid HTTP request methods.
var validMethods = map[string]bool{
	// The following methods are defined in RFC 2616:
	"OPTIONS": true,
	"GET":     true,
	"HEAD":    true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"TRACE":   true,
	"CONNECT": true,
	// The following methods are defined in RFC 5789:
	"PATCH": true,
}

// Return the canonical path for p, eliminating . and .. elements.
// Borrowed from net/http/server.go
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// match attempts to find a route that matches the given request.  Routes
// are evaluated in the order that they were created by NewRoute().
func match(req *http.Request, routes []*Route) *Route {
	for _, route := range routes {
		if !route.matchSchemes(req) ||
			!route.matchMethods(req) ||
			!route.matchHeaders(req) ||
			!route.matchHost(req) ||
			!route.matchPath(req) {
			continue
		}
		return route
	}
	return nil
}

// parseHost attempts to parse the provided host into a regular expression
// that can be used when matching routes.
func parseHost(host string) (*hostInfo, error) {
	// Empty hosts are not valid.
	if host == "" {
		return nil, fmt.Errorf(errEmptyHost)
	}

	pattern := bytes.NewBufferString("^")
	var depth, param, pos int
	for i := range host {
		switch host[i] {
		case '{':
			if depth++; depth == 1 {
				param = i
			}
		case '}':
			if depth--; depth == 0 {
				fmt.Fprintf(pattern, "%s(%s)", regexp.QuoteMeta(host[pos:param]), host[param+1:i])
				pos = i + 1
			} else if depth < 0 {
				// With properly formatted input, depth should never go below zero.
				return nil, fmt.Errorf(errUnevenBraces)
			}
		}
	}
	if depth != 0 {
		// At the end of the string, we're still inside a parameter brace.
		return nil, fmt.Errorf(errUnevenBraces)
	}

	if pos < len(host) {
		fmt.Fprint(pattern, regexp.QuoteMeta(host[pos:]))
	}
	pattern.WriteByte('$')

	re, err := regexp.Compile(pattern.String())
	if err != nil {
		return nil, err
	}

	return &hostInfo{
		rawHost: host,
		pattern: re,
	}, nil
}

// parsePath attempts to parse the provided path into a regular expression
// that can be used when matching routes.  It also creates a format string
// which can be used for printing a path with parameters filled in, as well
// as a slice of maps containing parameter names and regexp patterns.
func parsePath(path string, matchPrefix, matchSlashes bool) (*pathInfo, error) {
	// Empty paths are not valid.
	if path == "" {
		return nil, fmt.Errorf(errEmptyPath)
	}

	params := make([][]string, 0)
	fwdPattern := bytes.NewBufferString("^")
	revPattern := new(bytes.Buffer)
	var depth, param, pos int
	for i := range path {
		switch path[i] {
		case '{':
			if depth++; depth == 1 {
				param = i
			}
		case '}':
			if depth--; depth == 0 {
				nameVal := strings.SplitN(path[param+1:i], ":", 2)
				// Parameters must be named.
				if nameVal[0] == "" {
					return nil, fmt.Errorf(errParamNameNotDefined)
				}
				// Parameters must be unique per path.
				// FIXME: Do parameters really need to be unique?
				for p := 0; p < len(params); p++ {
					if params[p][0] == nameVal[0] {
						return nil, fmt.Errorf(errParamNameDefined, nameVal[0])
					}
				}

				if len(nameVal) < 2 {
					nameVal[1] = ""
				}
				if nameVal[1] == "" {
					nameVal[1] = "[^/]+"
				}
				subPath := path[pos:param]
				fmt.Fprintf(fwdPattern, "%s(%s)", regexp.QuoteMeta(subPath), nameVal[1])
				fmt.Fprintf(revPattern, "%s%%s", subPath)
				params = append(params, nameVal)
				pos = i + 1
			} else if depth < 0 {
				// With properly formatted input, depth should never go below zero.
				return nil, fmt.Errorf(errUnevenBraces)
			}
		}
	}
	if depth != 0 {
		// At the end of the string, we're still inside a parameter brace.
		return nil, fmt.Errorf(errUnevenBraces)
	}

	if pos < len(path) {
		fmt.Fprint(fwdPattern, regexp.QuoteMeta(path[pos:]))
		fmt.Fprint(revPattern, path[pos:])
	}

	if path != "/" && matchSlashes {
		if !strings.HasSuffix(path, "/") {
			fwdPattern.WriteByte('/')
		}
		fwdPattern.WriteByte('?')
	}
	if !matchPrefix {
		fwdPattern.WriteByte('$')
	}

	fwdRegexp, err := regexp.Compile(fwdPattern.String())
	if err != nil {
		return nil, err
	}

	return &pathInfo{
		rawPath:    path,
		fwdPattern: fwdRegexp,
		revPattern: revPattern.String(),
		params:     params,
	}, nil
}

// sliceContainsString checks to see if a string exists within a slice of
// strings.
func sliceContainsString(s []string, v string) bool {
	for _, element := range s {
		if v == element {
			return true
		}
	}
	return false
}

// sliceContainsStrings checks to see if all of the provided strings exist
// within a slice of strings.
func sliceContainsStrings(s, v []string) bool {
	for _, element := range v {
		if !sliceContainsString(s, element) {
			return false
		}
	}
	return true
}

// validateMethods takes a list of methods and verifies that they are
// supported.  It returns a properly formatted map on success, or an error if
// an unsupported method was provided.
func validateMethods(m ...string) (map[string]bool, error) {
	if len(m) == 0 {
		return nil, fmt.Errorf(errUnsupportedMethod, "")
	}
	methods := make(map[string]bool)
	for _, v := range m {
		v = strings.ToUpper(v)
		if !validMethods[v] {
			return nil, fmt.Errorf(errUnsupportedMethod, v)
		}
		methods[v] = true
	}
	return methods, nil
}

// validateSchemes takes a list of schemes and verifies that they are
// supported.  It returns a properly formatted map on success, or an error if
// an unsupported scheme was provided.
func validateSchemes(s ...string) (map[string]bool, error) {
	if len(s) == 0 {
		return nil, fmt.Errorf(errUnsupportedScheme, "")
	}
	schemes := make(map[string]bool)
	for _, v := range s {
		v = strings.ToLower(v)
		if v != "http" && v != "https" {
			return nil, fmt.Errorf(errUnsupportedScheme, v)
		}
		schemes[v] = true
	}
	return schemes, nil
}
