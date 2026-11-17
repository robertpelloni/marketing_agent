//go:build linux

package web

import "syscall"

// diskUsage reports root-filesystem usage so /health/detailed can show
// the disk cliff (shared VPS hit 91% from other tenants' backups in
// 2026-11) before Postgres runs out of space.
func diskUsage() (usedPercent int, freeBytes uint64, ok bool) {
	var st syscall.Statfs_t
	if err := syscall.Statfs("/", &st); err != nil {
		return 0, 0, false
	}
	total := st.Blocks * uint64(st.Bsize)
	free := st.Bavail * uint64(st.Bsize)
	if total == 0 {
		return 0, 0, false
	}
	return int((total - free) * 100 / total), free, true
}
