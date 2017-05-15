package pfsserver

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"GetFiles",
		"GET",
		"/api/v1/files",
		GetFiles,
	},
	Route{
		"PostFiles",
		"Post",
		"/api/v1/files",
		CreateFiles,
	},
}
