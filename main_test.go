package main_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/tbruyelle/keyring-compat"
)

func TestE2EGaiaV15(t *testing.T) {
	jsonSignerBin := filepath.Join(t.TempDir(), "json-signer")
	// Build json-signer bin
	err := exec.Command("go", "build", "-o", jsonSignerBin, ".").Run()
	if err != nil {
		t.Fatalf("can't build json-signer: %v", err)
	}
	gaiaNode := setupGaiaNode(t)
	gaiaCmd := gaiaNode.start(t)
	t.Cleanup(func() {
		gaiaCmd.Process.Kill()
	})

	testscript.Run(t, testscript.Params{
		Dir:      "testdata/gaiaV15",
		TestWork: true,
		Setup: func(env *testscript.Env) error {
			env.Setenv("GAIAD", gaiaNode.bin)
			env.Setenv("GAIA_HOME", gaiaNode.home)
			env.Setenv("JSONSIGNER", jsonSignerBin)
			env.Setenv("VAL1", gaiaNode.addrs.val1)
			env.Setenv("TEST1", gaiaNode.addrs.test1)
			env.Setenv("TEST2", gaiaNode.addrs.test2)
			return nil
		},
	})
}

type node struct {
	bin     string
	home    string
	chainID string
	addrs   struct {
		val1  string
		test1 string
		test2 string
	}
}

// TODO write generic setupNode with proper params
func setupGaiaNode(t *testing.T) node {
	dir := t.TempDir()
	gaiadBin := filepath.Join(dir, "gaiad")
	// Build gaiad bin
	err := exec.Command("go", "build", "-o", gaiadBin,
		"-modfile=testdeps/go.mod",
		"github.com/cosmos/gaia/v15/cmd/gaiad",
	).Run()
	if err != nil {
		t.Fatalf("can't build gaiad: %v", err)
	}
	n := node{
		bin:     gaiadBin,
		home:    filepath.Join(dir, "gaia"),
		chainID: "cosmos-dev",
	}
	keyringBackendFlag := "--keyring-backend=test"
	n.run(t, "init", "gaia-test", n.homeFlag(), "--chain-id="+n.chainID)
	n.run(t, "config", "chain-id", n.chainID, n.homeFlag())
	n.run(t, "keys", "add", "val1", n.homeFlag(), keyringBackendFlag)
	n.run(t, "keys", "add", "test1", n.homeFlag(), keyringBackendFlag)
	n.run(t, "keys", "add", "test2", n.homeFlag(), keyringBackendFlag)
	n.run(t, "genesis", "add-genesis-account", "val1", "1000000000stake", n.homeFlag(), keyringBackendFlag)
	n.run(t, "genesis", "add-genesis-account", "test1", "1000000000uatom", n.homeFlag(), keyringBackendFlag)
	n.run(t, "genesis", "gentx", "val1", "1000000000stake", n.homeFlag(), keyringBackendFlag)
	n.run(t, "genesis", "collect-gentxs", n.homeFlag())

	// fetch bech32 format of created addresses
	kr, err := keyring.New(
		keyring.BackendType("file"),
		filepath.Join(n.home, "keyring-test"),
		func(_ string) (string, error) { return "test", nil },
	)
	if err != nil {
		t.Fatalf("cannot access gaia keyring: %v", err)
	}
	n.addrs.val1 = getBech32Addr(t, kr, "val1.info", "cosmos")
	n.addrs.test1 = getBech32Addr(t, kr, "test1.info", "cosmos")
	n.addrs.test2 = getBech32Addr(t, kr, "test2.info", "cosmos")
	return n
}

func (n node) start(t *testing.T) *exec.Cmd {
	cmd := exec.Command(n.bin, "start", n.homeFlag(), "--minimum-gas-prices=100uatom")
	// cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		t.Fatalf("node start: %v", err)
	}
	go cmd.Wait()
	waitNodeReady(t, 10)
	return cmd
}

func (n node) homeFlag() string {
	return fmt.Sprintf("--home=%s", n.home)
}

func (n node) run(t *testing.T, args ...string) {
	cmd := exec.Command(n.bin, args...)
	// cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("node running '%s %s': %v", n.bin, strings.Join(args, " "), err)
	}
}

// waitNodeReady request the /status endpoint and ensures the
// sync_info.latest_block_hash is filled, meaning the node has started to
// produce blocks.
func waitNodeReady(t *testing.T, maxAttempts int) {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		t.Logf("wait node ready, attempt %d\n", attempt+1)
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
		err = json.NewDecoder(resp.Body).Decode(&status)
		if err == nil && status.Result.SyncInfo.LatestBlockHash != "" {
			// node ready
			return
		}
	}
	t.Fatalf("node not ready after %d attempts", maxAttempts)
}

func getBech32Addr(t *testing.T, kr keyring.Keyring, key, prefix string) string {
	k, err := kr.Get(key)
	if err != nil {
		t.Fatalf("cannot read key '%s' addr: %v", key, err)
	}
	addr, err := k.Bech32Address(prefix)
	if err != nil {
		t.Fatalf("cannot get bech32 format of key '%s': %v", key, err)
	}
	return addr
}
