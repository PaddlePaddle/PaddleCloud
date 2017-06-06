package paddlecloud

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"

	"github.com/google/subcommands"
)

// SimpleFileCmd define the subcommand of simple file operations
type SimpleFileCmd struct {
}

// Name is subcommands name
func (*SimpleFileCmd) Name() string { return "file" }

// Synopsis is subcommands synopsis
func (*SimpleFileCmd) Synopsis() string { return "Simple file operations." }

// Usage is subcommands Usage
func (*SimpleFileCmd) Usage() string {
	return `file [put|get] <src> <dst>:
	Options:
`
}

// SetFlags registers subcommands flags
func (p *SimpleFileCmd) SetFlags(f *flag.FlagSet) {
}

// Execute file ops
func (p *SimpleFileCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 || f.NArg() > 3 {
		f.Usage()
		return subcommands.ExitFailure
	}
	switch f.Arg(0) {
	case "put":
		err := putFile(f.Arg(1), f.Arg(2))
		if err != nil {
			fmt.Fprintf(os.Stderr, "put file error: %s\n", err)
			return subcommands.ExitFailure
		}
	case "get":
		err := getFile(f.Arg(1), f.Arg(2))
		if err != nil {
			fmt.Fprintf(os.Stderr, "get file error: %s\n", err)
			return subcommands.ExitFailure
		}
	default:
		f.Usage()
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

func putFile(src string, dest string) error {
	query := url.Values{}
	_, srcFile := path.Split(src)
	destDir, destFile := path.Split(dest)
	var destFullPath string
	if len(destFile) == 0 {
		destFullPath = path.Join(destDir, srcFile)
	} else {
		destFullPath = dest
	}
	query.Set("path", destFullPath)
	respStr, err := PostFile(config.ActiveConfig.Endpoint+"/api/v1/file/", src, query)
	if err != nil {
		return err
	}
	var respObj interface{}
	if err = json.Unmarshal(respStr, &respObj); err != nil {
		return err
	}
	// FIXME: Print an error if error message is not empty. Use response code instead
	errMsg := respObj.(map[string]interface{})["msg"].(string)
	if len(errMsg) > 0 {
		fmt.Fprintf(os.Stderr, "upload file error: %s\n", errMsg)
	}
	return nil
}

func getFile(src string, dest string) error {
	query := url.Values{}
	query.Set("path", src)
	req, err := makeRequestToken(config.ActiveConfig.Endpoint+"/api/v1/file/", "GET", nil, "", query)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.Status != HTTPOK {
		return errors.New("server error: " + resp.Status)
	}
	_, srcFile := path.Split(src)
	destDir, destFile := path.Split(dest)
	var destFullPath string
	if len(destFile) == 0 {
		destFullPath = path.Join(destDir, srcFile)
	} else {
		destFullPath = dest
	}
	if _, err = os.Stat(destFullPath); err == nil {
		return errors.New("file already exist: " + destFullPath)
	}
	out, err := os.Create(destFullPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
