package pfsserver

import (
	"net/http"

	pfsmod "github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
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
		"/" + pfsmod.RESTFilesPath,
		GetFilesHandler,
	},
	Route{
		"PostFiles",
		"POST",
		"/" + pfsmod.RESTFilesPath,
		PostFilesHandler,
	},
	Route{
		"DeleteFiles",
		"DELETE",
		"/" + pfsmod.RESTFilesPath,
		DeleteFilesHandler,
	},

	Route{
		"GetChunksMeta",
		"GET",
		"/" + pfsmod.RESTChunksPath,
		GetChunkMetaHandler,
	},
	Route{
		"GetChunksData",
		"GET",
		"/" + pfsmod.RESTChunksStoragePath,
		GetChunkHandler,
	},

	Route{
		"PostChunksData",
		"POST",
		"/" + pfsmod.RESTChunksStoragePath,
		PostChunkHandler,
	},
}
