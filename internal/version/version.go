package version

import (
	"errors"
	"runtime/debug"
)

var (
	Version string
)

type VCSInfo struct {
	VCS          string
	Revision     string
	Date         string
	TreeModified string
}

type Info struct {
	Version   string
	VCSInfo   VCSInfo
	GoVersion string
	Compiler  string
	Platform  string
}

func GetBuildInfo() (*Info, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("version: Error in ReadBuildInfo")
	}
	result := &Info{
		Version:   Version,
		GoVersion: info.GoVersion,
	}

	for _, setting := range info.Settings {
		switch setting.Key {
		case "-compiler":
			result.Compiler = setting.Value
		case "GOARCH":
			result.Platform += setting.Value
		case "GOOS":
			result.Platform = setting.Value + "/" + result.Platform
		case "vcs":
			result.VCSInfo.VCS = setting.Value
		case "vcs.revision":
			result.VCSInfo.Revision = setting.Value
		case "vcs.time":
			result.VCSInfo.Date = setting.Value
		case "vcs.modified":
			result.VCSInfo.TreeModified = setting.Value
		}
	}
	return result, nil
}
