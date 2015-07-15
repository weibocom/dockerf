package dflag

import (
	"os"

	mflag "github.com/docker/docker/pkg/mflag"
)

var DFlag = mflag.NewFlagSet(os.Args[0], mflag.ExitOnError)
