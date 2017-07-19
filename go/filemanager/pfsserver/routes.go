package pfsserver

import (
	"net/http"
)

// Route represents route struct.
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes is the a Route array
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
		"POST",
		"/api/v1/files",
		PostFilesHandler,
	},
	Route{
		"DeleteFiles",
		"DELETE",
		"/api/v1/files",
		DeleteFilesHandler,
	},

	Route{
		"GetChunksMeta",
		"GET",
		"/api/v1/chunks",
		GetChunkMetaHandler,
	},
	Route{
		"GetChunksData",
		"GET",
		"/api/v1/storage/chunks",
		GetChunkHandler,
	},

	Route{
		"PostChunksData",
		"POST",
		"/api/v1/storage/chunks",
		PostChunkHandler,
	},
}
