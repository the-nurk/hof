package cmds

import (
	"github.com/hofstadter-io/hofmod-cli/schema"
)

#FmtCommand: schema.#Command & {
	Name:  "fmt"
	Usage: "fmt [filepaths or globs]"
	Short: "format any code, manage formatters"
	Long:  #FmtRootHelp

	Args: [{
		Name: "files"
		Type: "[]string"
		Required: true
		Rest: true
		Help: "filepath or glob"
	}]

	_carg: [{
		Name:     "formatter"
		Type:     "string"
		Required: true
		Help:     "formatter name"
	}]

	Commands: [{
		Name:  "list"
		Usage: "list"
		Short: "list known formatters"
		Long:  Short
	}, {
		Name:  "info"
		Usage: "info"
		Short: "get formatter info"
		Long:  Short
		Args: _carg
	}, {
		Name:  "start"
		Usage: "start"
		Short: "start a formatter"
		Long:  Short
		Args: _carg
	}, {
		Name:  "stop"
		Usage: "stop"
		Short: "stop a formatter"
		Long:  Short
		Args: _carg
	}]

}

#FmtRootHelp: """
With hof fmt, you can
  1. format any language from a single tool
  2. run formatters as api servers for IDEs and hof
  3. manage the underlying formatter containers
"""
