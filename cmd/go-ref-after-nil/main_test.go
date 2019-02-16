package main_test

import (
	"os/exec"
	"strings"
	"syscall"
	"testing"
)

func TestCmd(t *testing.T) {
	const expectedOutput = `
Reference after nil check at ../../testdata/bad_conditional_print_nil_err.go:11:2
Reference after nil check at ../../testdata/bad_print_nil_err.go:10:2
Reference after nil check at ../../testdata/bad_return_nil_err.go:8:2
exit status 1
`
	cmd := exec.Command("go", "run", "main.go", "../../testdata")
	out, err := cmd.CombinedOutput()
	t.Log("output:\n", string(out))
	if err == nil {
		t.Errorf("expected check command %q to fail, but got success", cmd.Args)
	} else if exitErr, ok := err.(*exec.ExitError); !ok {
		t.Errorf("failed to run %q: %s", cmd.Args, err.Error())
	} else if ws, ok := exitErr.Sys().(syscall.WaitStatus); !ok {
		t.Errorf("unable to determine exit status of %q", cmd.Args)
	} else if ws.ExitStatus() != 1 {
		t.Errorf("expected %q to have exit status 1, got %d", cmd.Args, ws.ExitStatus())
	}
	if e, o := strings.TrimSpace(expectedOutput), strings.TrimSpace(string(out)); e != o {
		t.Error("output differs from expected")
		for i := 0; i < len(e) && i < len(o); i++ {
			if e[i] != o[i] {
				line := len(strings.Split(e[:i], "\n"))
				col := i - strings.LastIndex(e[:i], "\n")
				t.Logf("diff start at line %d, column %d, expected %q, got %q", line, col, e[i], o[i])
				break
			}
		}
	}
}
