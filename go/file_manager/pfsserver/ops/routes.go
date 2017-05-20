package pfsserver

import (
	//"github.com/gorilla/mux"
	"net/http"
)

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
		GetFilesHandler,
	},
	Route{
		"PostFiles",
		"Post",
		"/api/v1/files",
		PostFilesHandler,
	},
	Route{
		"GetChunksMeta",
		"GET",
		"/api/v1/chunks",
		GetChunksHandler,
	},
	Route{
		"GetChunksData",
		"Get",
		"/api/v1/storage/chunks",
		PostChunksHandler,
	},
}
