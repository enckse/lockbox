package internal

type (
	// Arguments options for operations.
	Arguments struct {
		Clip  bool
		Once  bool
		Short bool
		List  bool
		Multi bool
		Yes   bool
	}
)

// ParseArgs parses CLI arguments.
func ParseArgs(arg string) Arguments {
	args := Arguments{}
	args.Clip = arg == "-clip" || arg == "-c"
	args.Once = arg == "-once"
	args.Short = arg == "-short"
	args.List = arg == "-ls" || arg == "-list"
	args.Multi = arg == "-m" || arg == "-multi"
	args.Yes = arg == "-yes"
	return args
}
