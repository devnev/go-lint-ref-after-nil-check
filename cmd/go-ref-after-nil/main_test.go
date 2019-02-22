package main_test

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

const (
	testdataPath = "../../testdata"
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
	cmd := exec.Command("go", "run", "main.go", testdataPath)
	out, err := cmd.CombinedOutput()
	t.Logf("output:\n%s", string(out))
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
	cmd := exec.Command("go", "run", "main.go", "-machine", "-exit-code", testdataPath)
	out, err := cmd.CombinedOutput()
	t.Logf("output:\n%s", string(out))
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

func TestCmd_fix(t *testing.T) {
	const expectedOutput = `
diff -ru '--label=testdata' '--label=fixed' testdata fixed
--- testdata
+++ fixed
@@ -9,6 +9,6 @@
 		return
 	}
 	if !expected {
-		fmt.Println(err)
+		fmt.Println(nil)
 	}
 }
diff -ru '--label=testdata' '--label=fixed' testdata fixed
--- testdata
+++ fixed
@@ -7,5 +7,5 @@
 	if err != nil {
 		return
 	}
-	fmt.Println(err)
+	fmt.Println(nil)
 }
diff -ru '--label=testdata' '--label=fixed' testdata fixed
--- testdata
+++ fixed
@@ -5,5 +5,5 @@
 	if err != nil {
 		return err
 	}
-	return err
+	return nil
 }
`
	path, err := ioutil.TempDir(".", "fixtest")
	if err != nil {
		t.Fatalf("Failed to create directory for fixing files: %s", err.Error())
	}
	defer func() {
		err := os.RemoveAll(path)
		if err != nil {
			t.Errorf("Faield to clean up temp dir: %s", err.Error())
		}
	}()
	list, err := ioutil.ReadDir(testdataPath)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %s", path, err.Error())
	}
	for _, fi := range list {
		srcp := filepath.Join(testdataPath, fi.Name())
		src, err := os.Open(srcp)
		if err != nil {
			t.Fatalf("Failed to open source file %s", srcp)
		}
		dstp := filepath.Join(path, fi.Name())
		dst, err := os.Create(dstp)
		if err != nil {
			t.Errorf("failed to create %s", dstp)
		}
		_, err = io.Copy(dst, src)
		if err != nil {
			t.Errorf("Failed to copy content of %s to %s", srcp, dstp)
		}
	}
	cmd := exec.Command("go", "run", "main.go", "-fix", path)
	out, err := cmd.CombinedOutput()
	t.Logf("output:\n%s", string(out))
	if err != nil {
		t.Errorf("Failed to run %q: %s", cmd.Args, err.Error())
	}
	cmd = exec.Command("diff", "-ru", "--label=testdata", testdataPath, "--label=fixed", path)
	out, err = cmd.CombinedOutput()
	t.Logf("diff output:\n%s", string(out))
	if err != nil {
		t.Logf("diff error: %s", err.Error())
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
