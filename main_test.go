package main_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/tbruyelle/keyring-compat"
)

func TestE2EGaiaV15(t *testing.T) {
	runE2ETest(t, "testdata/gaiaV15", setupGaiaNode(t))
}

func TestE2EGovgenV1(t *testing.T) {
	runE2ETest(t, "testdata/govgenV1", setupGovgenNode(t))
}

func runE2ETest(t *testing.T, dir string, node node) {
	jsonSignerBin := filepath.Join(t.TempDir(), "json-signer")
	// Build json-signer bin
	err := exec.Command("go", "build", "-o", jsonSignerBin, ".").Run()
	if err != nil {
		t.Fatalf("can't build json-signer: %v", err)
	}
	cmd := node.start(t)
	t.Cleanup(func() {
		cmd.Process.Kill()
	})

	testscript.Run(t, testscript.Params{
		Dir: dir,
		Setup: func(env *testscript.Env) error {
			nodeBin := node.bin
			if alternateBin := os.Getenv("NODE_BIN"); alternateBin != "" {
				t.Logf("'%s' bin overrided by env '%s'", nodeBin, alternateBin)
				nodeBin = alternateBin
			}
			env.Setenv("NODE_BIN", nodeBin)
			env.Setenv("NODE_HOME", node.home)
			env.Setenv("CHAINID", node.chainID)
			env.Setenv("JSONSIGNER", jsonSignerBin)
			env.Setenv("VAL1", node.addrs.val1)
			env.Setenv("VAL2", node.addrs.val2)
			env.Setenv("TEST1", node.addrs.test1)
			env.Setenv("TEST2", node.addrs.test2)
			env.Setenv("TEST3", node.addrs.test3)
			env.Setenv("MULTISIG", node.addrs.multisig)
			return nil
		},
	})
}

type node struct {
	bin     string
	home    string
	chainID string
	addrs   struct {
		val1     string
		val2     string
		test1    string
		test2    string
		test3    string
		multisig string
	}
}

func setupGovgenNode(t *testing.T) node {
	dir := t.TempDir()
	govgendBin := filepath.Join(dir, "govgend")
	// Build gaiad bin
	err := exec.Command("go", "build", "-o", govgendBin,
		"-modfile=testdeps/govgenV1/go.mod",
		"github.com/atomone-hub/govgen/cmd/govgend",
	).Run()
	if err != nil {
		t.Fatalf("can't build govgend: %v", err)
	}
	n := node{
		bin:     govgendBin,
		home:    filepath.Join(dir, "govgen"),
		chainID: "govgen-dev",
	}
	keyringBackendFlag := "--keyring-backend=test"
	chainIDFlag := "--chain-id=" + n.chainID
	n.run(t, "init", "govgen-test", n.homeFlag(), chainIDFlag)
	n.run(t, "config", "chain-id", n.chainID, n.homeFlag())
	n.run(t, "keys", "add", "val1", n.homeFlag(), keyringBackendFlag)
	n.run(t, "keys", "add", "val2", n.homeFlag(), keyringBackendFlag)
	n.run(t, "keys", "add", "test1", n.homeFlag(), keyringBackendFlag)
	n.run(t, "keys", "add", "test2", n.homeFlag(), keyringBackendFlag)
	n.run(t, "keys", "add", "test3", n.homeFlag(), keyringBackendFlag)
	n.run(t, "keys", "add", "test1-test2-multisig", "--multisig=test1,test2", "--multisig-threshold=2", n.homeFlag(), keyringBackendFlag)
	n.run(t, "add-genesis-account", "val1", "1000000000stake", n.homeFlag(), keyringBackendFlag)
	n.run(t, "add-genesis-account", "test1", "1000000000stake", n.homeFlag(), keyringBackendFlag)
	n.run(t, "add-genesis-account", "test2", "1000000000stake", n.homeFlag(), keyringBackendFlag)
	n.run(t, "add-genesis-account", "val2", "1000000000stake", n.homeFlag(), keyringBackendFlag)
	n.run(t, "add-genesis-account", "test1-test2-multisig", "1000000000stake", n.homeFlag(), keyringBackendFlag)
	n.run(t, "gentx", "val1", "1000000000stake", n.homeFlag(), chainIDFlag, keyringBackendFlag)
	n.run(t, "collect-gentxs", n.homeFlag())

	// fetch bech32 format of created addresses
	kr, err := keyring.New(
		keyring.BackendType("file"),
		filepath.Join(n.home, "keyring-test"),
		func(_ string) (string, error) { return "test", nil },
	)
	if err != nil {
		t.Fatalf("cannot access gaia keyring: %v", err)
	}
	n.addrs.val1 = getBech32Addr(t, kr, "val1.info", "govgenvaloper")
	n.addrs.val2 = getBech32Addr(t, kr, "val2.info", "govgenvaloper")
	n.addrs.test1 = getBech32Addr(t, kr, "test1.info", "govgen")
	n.addrs.test2 = getBech32Addr(t, kr, "test2.info", "govgen")
	n.addrs.test3 = getBech32Addr(t, kr, "test3.info", "govgen")
	n.addrs.multisig = getBech32Addr(t, kr, "test1-test2-multisig.info", "govgen")
	return n
}

func setupGaiaNode(t *testing.T) node {
	dir := t.TempDir()
	gaiadBin := filepath.Join(dir, "gaiad")
	// Build gaiad bin
	err := exec.Command("go", "build", "-o", gaiadBin,
		"-modfile=testdeps/gaiaV15/go.mod",
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
	n.run(t, "keys", "add", "val2", n.homeFlag(), keyringBackendFlag)
	n.run(t, "keys", "add", "test1", n.homeFlag(), keyringBackendFlag)
	n.run(t, "keys", "add", "test2", n.homeFlag(), keyringBackendFlag)
	n.run(t, "keys", "add", "test3", n.homeFlag(), keyringBackendFlag)
	n.run(t, "keys", "add", "test1-test2-multisig", "--multisig=test1,test2", "--multisig-threshold=2", n.homeFlag(), keyringBackendFlag)
	n.run(t, "genesis", "add-genesis-account", "val1", "1000000000stake", n.homeFlag(), keyringBackendFlag)
	n.run(t, "genesis", "add-genesis-account", "test1", "1000000000stake", n.homeFlag(), keyringBackendFlag)
	n.run(t, "genesis", "add-genesis-account", "test2", "1000000000stake", n.homeFlag(), keyringBackendFlag)
	n.run(t, "genesis", "add-genesis-account", "val2", "1000000000stake", n.homeFlag(), keyringBackendFlag)
	n.run(t, "genesis", "add-genesis-account", "test1-test2-multisig", "1000000000stake", n.homeFlag(), keyringBackendFlag)
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
	n.addrs.val1 = getBech32Addr(t, kr, "val1.info", "cosmosvaloper")
	n.addrs.val2 = getBech32Addr(t, kr, "val2.info", "cosmosvaloper")
	n.addrs.test1 = getBech32Addr(t, kr, "test1.info", "cosmos")
	n.addrs.test2 = getBech32Addr(t, kr, "test2.info", "cosmos")
	n.addrs.test3 = getBech32Addr(t, kr, "test3.info", "cosmos")
	n.addrs.multisig = getBech32Addr(t, kr, "test1-test2-multisig.info", "cosmos")
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
	bz, err := exec.Command(n.bin, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("node running '%s %s': %v\n%s", n.bin, strings.Join(args, " "), err, string(bz))
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
