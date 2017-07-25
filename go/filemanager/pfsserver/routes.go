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
		"/api/v1/pfs/files",
		GetFilesHandler,
	},
	Route{
		"PostFiles",
		"POST",
		"/api/v1/pfs/files",
		PostFilesHandler,
	},
	Route{
		"DeleteFiles",
		"DELETE",
		"/api/v1/pfs/files",
		DeleteFilesHandler,
	},

	Route{
		"GetChunksMeta",
		"GET",
		"/api/v1/pfs/chunks",
		GetChunkMetaHandler,
	},
	Route{
		"GetChunksData",
		"GET",
		"/api/v1/pfs/storage/chunks",
		GetChunkHandler,
	},

	Route{
		"PostChunksData",
		"POST",
		"/api/v1/pfs/storage/chunks",
		PostChunkHandler,
	},
}
