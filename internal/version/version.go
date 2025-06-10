package version

import "fmt"

var version string

func Name() string {
	return "QuakeWorld Broadcast Service"
}

func Short() string {
	return version
}

func Long() string {
	return fmt.Sprintf("%s %s", Name(), Short())
}
