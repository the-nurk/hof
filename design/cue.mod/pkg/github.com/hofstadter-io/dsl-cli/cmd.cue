package dsl_cli

Command : {
  Aliases:  [string]

  Usage:    string
  Short:    string
  Long:     string

  PersistentPrerun:   bool
  Prerun:             bool
  OmitRun:            bool
  Postrun:            bool
  PersistentPostrun:  bool

  Pflags:    [Flag]
  Flags:     [Flag]
  Args:      [Arg]
  Commands:  [Command]
}

