package main_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/tbruyelle/keyring-compat"
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
			"fillenvs": func(ts *testscript.TestScript, neg bool, args []string) {
				// Fill $TEST1 with bech32 address
				kr, err := keyring.New(
					keyring.BackendType("file"),
					filepath.Join(ts.Getenv("WORK"), "gaiad", "keyring-test"),
					func(_ string) (string, error) { return "test", nil },
				)
				if err != nil {
					ts.Fatalf(err.Error())
				}
				k, err := kr.Get("test1.info")
				if err != nil {
					ts.Fatalf(err.Error())
				}
				addr, err := k.Bech32Address("cosmos")
				if err != nil {
					ts.Fatalf(err.Error())
				}
				ts.Setenv("TEST1", addr)
			},
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
