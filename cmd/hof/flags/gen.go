package flags

type GenFlagpole struct {
	Stats     bool
	Generator []string
	Template  []string
	Partial   []string
	Diff3     bool
	Watch     []string
	WatchXcue []string
}

var GenFlags GenFlagpole
