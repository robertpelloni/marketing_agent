//go:build !linux

package web

// diskUsage is Linux-only (prod); other platforms report unknown.
func diskUsage() (usedPercent int, freeBytes uint64, ok bool) {
	return 0, 0, false
}
