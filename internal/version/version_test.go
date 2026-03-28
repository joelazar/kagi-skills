package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestInfo(t *testing.T) {
	info := Info()

	if !strings.Contains(info, Version) {
		t.Errorf("Info() should contain version %q, got: %s", Version, info)
	}
	if !strings.Contains(info, runtime.Version()) {
		t.Errorf("Info() should contain Go version, got: %s", info)
	}
	if !strings.Contains(info, "commit:") {
		t.Errorf("Info() should contain commit label, got: %s", info)
	}
}
