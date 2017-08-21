package pfsmodules

// TODO
// need a custom error type?

const (
	// StatusFileNotFound is a error string of that there is no file or directory.
	StatusFileNotFound = "no such file or directory"
	// StatusDirectoryNotAFile is a error string of that the destination should be a file.
	StatusDirectoryNotAFile = "should be a file not a directory"
	// StatusCopyFromLocalToLocal is a error string of that this system does't support copy local to local.
	StatusCopyFromLocalToLocal = "don't support copy local to local"
	// StatusDestShouldBeDirectory is a error string of that destination shoule be a directory.
	StatusDestShouldBeDirectory = "dest should be a directory"
	// StatusOnlySupportFiles is a error string of that the system only support upload or download files not directories.
	StatusOnlySupportFiles = "only support upload or download files not directories"
	// StatusBadFileSize is a error string of that the file size is no valid.
	StatusBadFileSize = "bad file size"
	// StatusDirectoryAlreadyExist is a error string of that the directory is already exist.
	StatusDirectoryAlreadyExist = "directory already exist"
	// StatusBadChunkSize is a error string of that the chunksize is error.
	StatusBadChunkSize = "chunksize error"
	// StatusShouldBePfsPath is a error string of that a path should be a pfs path.
	StatusShouldBePfsPath = "should be pfs path"
	// StatusNotEnoughArgs is a error string of that there is not enough arguments.
	StatusNotEnoughArgs = "not enough arguments"
	// StatusInvalidArgs is a error string of that arguments are not valid.
	StatusInvalidArgs = "invalid arguments"
	// StatusUnAuthorized is a error string of that what you request should have authorization.
	StatusUnAuthorized = "what you request is unauthorized"
	// StatusJSONErr is a error string of that the system parses json error.
	StatusJSONErr = "parse json error"
	// StatusCannotDelDirectory is a error string of that what you input can't delete a directory.
	StatusCannotDelDirectory = "can't del directory"
	// StatusAlreadyExist is a error string of that the destination is already exist.
	StatusAlreadyExist = "already exist"
	// StatusBadPath is a error string of that the form of path is not correct.
	StatusBadPath = "the path should be in format eg:/pf/datacentername/"

	// StatusFileEOF is a status string indicates that the file reaches end
	StatusFileEOF = "this file reaches end"
)
