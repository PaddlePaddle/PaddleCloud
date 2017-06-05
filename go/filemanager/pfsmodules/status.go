package pfsmodules

// TODO
// need a custom error type?

const (
	StatusFileNotFound          = "no such file or directory"
	StatusDirectoryNotAFile     = "should be a file not a directory"
	StatusCopyFromLocalToLocal  = "don't support copy local to local"
	StatusDestShouldBeDirectory = "dest should be a directory"
	StatusOnlySupportFiles      = "only support upload or download files not directories"
	StatusBadFileSize           = "bad file size"
	StatusDirectoryAlreadyExist = "directory already exist"
	StatusBadChunkSize          = "chunksize error"
	StatusShouldBePfsPath       = "should be pfs path"
	StatusNotEnoughArgs         = "not enough arguments"
	StatusInvalidArgs           = "invalid arguments"
	StatusUnAuthorized          = "what you request is unauthorized"
	StatusJSONErr               = "parse json error"
	StatusCannotDelDirectory    = "can't del directory"
	StatusAlreadyExist          = "already exist"
	StatusBadPath               = "the path should be in format eg:/pf/datacentername/"
)
