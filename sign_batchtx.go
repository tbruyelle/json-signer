package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"

	"github.com/tbruyelle/keyring-compat"
)

func batchSignTx(kr keyring.Keyring, fileLoc, signer, chainID, account, sequence string) ([]Tx, [][]byte, error) {
	// Read the file
	txs, err := readTxs(fileLoc)
	if err != nil {
		return nil, nil, err
	}

	// Sign each tx
	seq, err := strconv.ParseInt(sequence, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	var bytesSigned [][]byte
	for i, tx := range txs {
		sequence = strconv.FormatInt(seq+int64(i), 10)

		tx1, bytesToSign, err := signTx(tx, kr, signer, chainID, account, sequence)
		if err != nil {
			return nil, nil, err
		}
		txs[i] = tx1
		bytesSigned = append(bytesSigned, bytesToSign)
	}

	return txs, bytesSigned, nil
}

func readTxs(fileLoc string) ([]Tx, error) {
	f, err := os.Open(fileLoc)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Create a new scanner for the file
	scanner := bufio.NewScanner(f)

	// Read and append each line of the file
	var txs []Tx
	for scanner.Scan() {
		line := scanner.Text()
		var tx Tx
		if err := json.Unmarshal([]byte(line), &tx); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}

	// Check for errors during the scan
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return txs, nil
}
