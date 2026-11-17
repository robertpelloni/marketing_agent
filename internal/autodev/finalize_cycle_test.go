package autodev

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFinalizeCycleSkipsVersionWritesInTestMode(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	t.Setenv("GO_TEST_MODE", "true")

	o := &Orchestrator{}
	o.finalizeCycle(context.Background(), &Task{Description: "Test integration task"})

	for _, name := range []string{"VERSION", "VERSION.md", "CHANGELOG.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Errorf("finalizeCycle must not write %s in GO_TEST_MODE (repo pollution), stat err: %v", name, err)
		}
	}
}

func TestFinalizeCycleWritesOutsideTestMode(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	t.Setenv("GO_TEST_MODE", "")

	o := &Orchestrator{}
	o.finalizeCycle(context.Background(), &Task{Description: "Real task"})

	if _, err := os.Stat(filepath.Join(dir, "VERSION")); err != nil {
		t.Errorf("expected VERSION to be written outside test mode: %v", err)
	}
}
