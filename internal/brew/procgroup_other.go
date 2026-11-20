//go:build !unix

package brew

import "os/exec"

// setProcessGroup is a no-op where process groups are not available: a
// cancelled context kills brew alone, as before.
func setProcessGroup(cmd *exec.Cmd) {}
