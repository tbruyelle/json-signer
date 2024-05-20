package main_test

import (
	"os/exec"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestScripts(t *testing.T) {
	// Build json-signer bin
	err := exec.Command("go", "build", "-o=/tmp/json-signer", ".").Run()
	if err != nil {
		t.Fatalf("can't build json-signer: %v", err)
	}
	// Build gaiad bin
	err = exec.Command("go", "build", "-o=/tmp/gaiad",
		"-modfile=testdeps/go.mod",
		"github.com/cosmos/gaia/v15/cmd/gaiad",
	).Run()
	if err != nil {
		t.Fatalf("can't build json-signer: %v", err)
	}

	testscript.Run(t, testscript.Params{
		Dir:      "testdata",
		TestWork: true,
		/*
			Setup: func(env *testscript.Env) error {
				if err := runGaiad(env.Cd, "init", "gaia-test"); err != nil {
					return err
				}
				if err := runGaiad(env.Cd, strings.Fields("keys add test1 --keyring-backend=test")...); err != nil {
					return err
				}
				return nil
			},
		*/
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"gaiad": func(ts *testscript.TestScript, neg bool, args []string) {
				tsExec(ts, neg, "/tmp/gaiad", args)
			},
			"json-signer": func(ts *testscript.TestScript, neg bool, args []string) {
				tsExec(ts, neg, "/tmp/json-signer", args)
			},
		},
	})
}

func tsExec(ts *testscript.TestScript, neg bool, cmd string, args []string) {
	err := ts.Exec(cmd, args...)
	if err != nil {
		ts.Logf("%s command error: %+v", cmd, err)
	}

	commandSucceeded := (err == nil)
	successExpected := !neg

	// Compare the command's success status with the expected outcome.
	if commandSucceeded != successExpected {
		ts.Fatalf("unexpected %s command outcome (err=%t expected=%t)", cmd, commandSucceeded, successExpected)
	}
	if err != nil {
		// TODO handle neg param
		ts.Fatalf(err.Error())
	}
}
