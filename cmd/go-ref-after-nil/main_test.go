package main_test

import (
	"os/exec"
	"strings"
	"syscall"
	"testing"
)

func TestCmd(t *testing.T) {
	const expectedOutput = `
Reference after nil check at ../../testdata/bad_conditional_print_nil_err.go:12:15
		fmt.Println(err)
Reference after nil check at ../../testdata/bad_print_nil_err.go:10:14
	fmt.Println(err)
Reference after nil check at ../../testdata/bad_return_nil_err.go:8:9
	return err
`
	cmd := exec.Command("go", "run", "main.go", "../../testdata")
	out, err := cmd.CombinedOutput()
	t.Log("output:\n", string(out))
	if err != nil {
		t.Errorf("expected check command %q to succeed, got %s", cmd.Args, err.Error())
	}
	testOutput(t, expectedOutput, string(out))
}

func TestCmd_machineOutput(t *testing.T) {
	const expectedOutput = `
../../testdata/bad_conditional_print_nil_err.go:12:15
../../testdata/bad_print_nil_err.go:10:14
../../testdata/bad_return_nil_err.go:8:9
exit status 2
`
	cmd := exec.Command("go", "run", "main.go", "-machine", "-exit-code", "../../testdata")
	out, err := cmd.CombinedOutput()
	t.Log("output:\n", string(out))
	if err == nil {
		t.Errorf("expected check command %q to fail, but got success", cmd.Args)
	} else if exitErr, ok := err.(*exec.ExitError); !ok {
		t.Errorf("failed to run %q: %s", cmd.Args, err.Error())
	} else if ws, ok := exitErr.Sys().(syscall.WaitStatus); !ok {
		t.Errorf("unable to determine exit status of %q", cmd.Args)
	} else if ws.ExitStatus() != 1 {
		// note that this is checking `go run`'s exit status, not the linters. the linter's exit status is in the output.
		t.Errorf("expected %q to have exit status 1, got %d", cmd.Args, ws.ExitStatus())
	}
	testOutput(t, expectedOutput, string(out))
}

func testOutput(t *testing.T, expected, output string) {
	expected, output = strings.TrimSpace(expected), strings.TrimSpace(output)
	if expected == output {
		return
	}
	t.Error("output differs from expected")
	for i := 0; i < len(expected) && i < len(output); i++ {
		if expected[i] != output[i] {
			line := len(strings.Split(expected[:i], "\n"))
			col := i - strings.LastIndex(expected[:i], "\n")
			t.Logf("diff start at line %d, column %d, expected %q, got %q", line, col, expected[i], output[i])
			break
		}
	}
}
