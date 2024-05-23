package main_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/tbruyelle/keyring-compat"
)

func TestE2E(t *testing.T) {
	var (
		dir           = t.TempDir()
		gaiadBin      = filepath.Join(dir, "gaiad")
		jsonSignerBin = filepath.Join(dir, "json-signer")
	)
	// Build json-signer bin
	err := exec.Command("go", "build", "-o", jsonSignerBin, ".").Run()
	if err != nil {
		t.Fatalf("can't build json-signer: %v", err)
	}
	// Build gaiad bin
	err = exec.Command("go", "build", "-o", gaiadBin,
		"-modfile=testdeps/go.mod",
		"github.com/cosmos/gaia/v15/cmd/gaiad",
	).Run()
	if err != nil {
		t.Fatalf("can't build json-signer: %v", err)
	}

	testscript.Run(t, testscript.Params{
		Dir:      "testdata",
		TestWork: true,
		Setup: func(env *testscript.Env) error {
			env.Setenv("GAIAD", gaiadBin)
			env.Setenv("JSONSIGNER", jsonSignerBin)
			return nil
		},
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
				// Wait node ready
				waitNodeReady(ts, 10)
			},
		},
	})
}

// waitNodeReady request the /status endpoint and ensures the
// sync_info.latest_block_hash is filled, meaning the node has started to
// produce blocks.
func waitNodeReady(ts *testscript.TestScript, maxAttempts int) {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		fmt.Printf("wait node ready, attempt %d\n", attempt+1)
		time.Sleep(time.Second)
		resp, err := http.Get("http://localhost:26657/status")
		if err != nil {
			continue
		}
		var status struct {
			Result struct {
				SyncInfo struct {
					LatestBlockHash string `json:"latest_block_hash"`
				} `json:"sync_info"`
			} `json:"result"`
		}
		// var b bytes.Buffer
		// io.Copy(&b, resp.Body)
		// fmt.Println("OUT", b.String())
		err = json.NewDecoder(resp.Body).Decode(&status)
		if err == nil && status.Result.SyncInfo.LatestBlockHash != "" {
			// node ready
			return
		}
	}
	ts.Fatalf("node not ready after %d attempts", maxAttempts)
}
