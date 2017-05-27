package pfsserver

import (
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
		"POST",
		"/api/v1/files",
		PostFilesHandler,
	},

	Route{
		"GetChunksMeta",
		"GET",
		"/api/v1/chunks",
		GetChunksMetaHandler,
	},
	Route{
		"GetChunksData",
		"GET",
		"/api/v1/storage/chunks",
		GetChunksHandler,
	},

	Route{
		"PostChunksData",
		"POST",
		"/api/v1/storage/chunks",
		PostChunksHandler,
	},
}
