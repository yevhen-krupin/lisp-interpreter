// based on https://github.com/carbocation/genomisc/blob/master/compileinfo/compileinfo.go
package main

import (
	"runtime/debug"
)

type CompileInfo struct {
	Package    string
	GoVersion  string
	Commit     string
	CommitTime string
	Modified   bool
	Version    string
}

func (CompileInfo) Get() CompileInfo {
	out := CompileInfo{}

	z, ok := debug.ReadBuildInfo()
	if !ok {
		return out
	}

	out.GoVersion = z.GoVersion
	out.Package = z.Path
	out.Version = z.Main.Version
	for _, s := range z.Settings {
		switch s.Key {
		case "vcs.revision":
			out.Commit = s.Value
		case "vcs.time":
			out.CommitTime = s.Value
		case "vcs.modified":
			out.Modified = s.Value == "true"
		}
	}

	return out
}
