package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update .obj.ascii.golden files")

func runTest(t *testing.T, in, out, gold string) {

	_, memory := processAssembly(in)
	dumpASCII(out, memory)

	generatedBytes, err := ioutil.ReadFile(gold)
	if err != nil {
		t.Error(fmt.Sprintf("%+v", err))
		return
	}
	generated := string(generatedBytes)

	expectedBytes, err := ioutil.ReadFile(gold)
	if err != nil {
		t.Error(fmt.Sprintf("%+v", err))
		return
	}
	expected := string(expectedBytes)

	if generated != expected {
		if *update {
			if in != out {
				if err := ioutil.WriteFile(out, []byte(generated), 0666); err != nil {
					t.Error(fmt.Sprintf("%+v", err))
				}
				return
			}
			// in == out: don't accidentally destroy input
			t.Errorf("WARNING: -update did not rewrite input file %s", in)
		}

		t.Errorf("(lc3asm2obj %s) != %s (see %s.lc3asm2obj)", in, out, in)
		d, err := diff(expected, generated, in)
		if err == nil {
			t.Errorf("%s", d)
		}
	}
	if err := ioutil.WriteFile(in+".lc3asm2obj", []byte(generated), 0666); err != nil {
		t.Error(fmt.Sprintf("%+v", err))
	}
}

func Testlc3asm2obj(t *testing.T) {

	// determine input files
	match, err := filepath.Glob("testdata/*")
	t.Logf("%+v\n", match)
	if err != nil {
		t.Fatal(fmt.Sprintf("MATCH:%+v\n", err))
	}

	for _, in := range match {
		var out string
		var gold string
		if strings.HasSuffix(in, ".asm") {
			out = in[:len(in)-len(".asm")] + ".obj.ascii"
		}
		if strings.HasSuffix(in, ".asm") {
			gold = in[:len(in)-len(".asm")] + ".obj.ascii.golden"
		}

		runTest(t, in, out, gold)
	}
}

func diff(b1, b2 string, filename string) (data string, err error) {
	f1, err := writeTempFile("", "svfmt", b1)
	if err != nil {
		return
	}
	defer os.Remove(f1)

	f2, err := writeTempFile("", "svfmt", b2)
	if err != nil {
		return
	}
	defer os.Remove(f2)

	cmd := "diff"

	dataBytes, err := exec.Command(cmd, "-u", f1, f2).CombinedOutput()
	data = string(dataBytes)
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		return replaceTempFilename(data, filename)
	}
	return
}

// replaceTempFilename replaces temporary filenames in diff with actual one.
//
// --- /tmp/svfmt3161453762017-02-03 19:13:00.280468375 -0500
// +++ /tmp/svfmt6178828152017-02-03 19:13:00.280468375 -0500
// ...
// ->
// --- path/to/file.go.orig2017-02-03 19:13:00.280468375 -0500
// +++ path/to/file.go2017-02-03 19:13:00.280468375 -0500
// ...
func replaceTempFilename(diff string, filename string) (string, error) {
	bs := strings.SplitN(diff, "\n", 3)
	if len(bs) < 3 {
		return "", fmt.Errorf("got unexpected diff for %s", filename)
	}
	// Preserve timestamps.
	var t0, t1 string
	if i := strings.LastIndexByte(bs[0], '\t'); i != -1 {
		t0 = bs[0][i:]
	}
	if i := strings.LastIndexByte(bs[1], '\t'); i != -1 {
		t1 = bs[1][i:]
	}
	// Always print filepath with slash separator.
	f := filepath.ToSlash(filename)
	bs[0] = fmt.Sprintf("--- %s%s", f+".orig", t0)
	bs[1] = fmt.Sprintf("+++ %s%s", f, t1)
	return strings.Join(bs, "\n"), nil
}

func writeTempFile(dir, prefix string, data string) (string, error) {
	file, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(data)
	if err1 := file.Close(); err == nil {
		err = err1
	}
	if err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}
