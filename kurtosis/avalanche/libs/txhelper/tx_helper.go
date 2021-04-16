package txhelper

import (
	"fmt"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
)

// CreateSingleUTXOTx returns a transaction spending an individual utxo owned by [privateKey]
func CreateSingleUTXOTx(utxo *avax.UTXO, inputAmount, outputAmount uint64, address ids.ShortID, privateKey *crypto.PrivateKeySECP256K1R, codec codec.Manager) (*avm.Tx, error) {
	keys := [][]*crypto.PrivateKeySECP256K1R{{privateKey}}
	outs := []*avax.TransferableOutput{
		{
			Asset: avax.Asset{ID: constants.AvaxAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: outputAmount,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{address},
				},
			},
		},
	}

	transferableIn := interface{}(&secp256k1fx.TransferInput{
		Amt: inputAmount,
		Input: secp256k1fx.Input{
			SigIndices: []uint32{0},
		},
	})

	ins := []*avax.TransferableInput{
		{
			UTXOID: utxo.UTXOID,
			Asset:  avax.Asset{ID: constants.AvaxAssetID},
			In:     transferableIn.(avax.TransferableIn),
		},
	}

	tx := &avm.Tx{UnsignedTx: &avm.BaseTx{BaseTx: avax.BaseTx{
		NetworkID:    constants.NetworkID,
		BlockchainID: constants.XChainID,
		Outs:         outs,
		Ins:          ins,
	}}}

	if err := tx.SignSECP256K1Fx(codec, keys); err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateConsecutiveTransactions returns a string of [numTxs] sending [utxo] back and forth
// assumes that [privateKey] is the sole owner of [utxo]
func CreateConsecutiveTransactions(utxo *avax.UTXO, numTxs, amount, txFee uint64, privateKey *crypto.PrivateKeySECP256K1R, codec codec.Manager) ([][]byte, []ids.ID, error) {
	if numTxs*txFee > amount {
		return nil, nil, fmt.Errorf("Insufficient starting funds to send %v transactions with a txFee of %v", numTxs, txFee)
	}

	address := privateKey.PublicKey().Address()
	txBytes := make([][]byte, numTxs)
	txIDs := make([]ids.ID, numTxs)

	inputAmount := amount
	outputAmount := amount - txFee
	for i := uint64(0); i < numTxs; i++ {
		tx, err := CreateSingleUTXOTx(utxo, inputAmount, outputAmount, address, privateKey, codec)
		if err != nil {
			return nil, nil, err
		}
		txBytes[i] = tx.Bytes()
		txIDs[i] = tx.ID()
		utxo = tx.UTXOs()[0]
		inputAmount -= txFee
		outputAmount -= txFee
	}

	return txBytes, txIDs, nil
}

// CreateIndependentBurnTxs creates a list of transactions spending each utxo in [utxos] and sending [utxoAmount] - [txFee] back
// to [privateKey]
// Assumes that each utxo has a sufficient amount to pay the transaction fee.
func CreateIndependentBurnTxs(utxos []*avax.UTXO, utxoAmount, txFee uint64, privateKey *crypto.PrivateKeySECP256K1R, codec codec.Manager) ([][]byte, []ids.ID, error) {
	var txBytes [][]byte
	var txIDs []ids.ID

	// logrus.Infof("CreateIndependentBurnTxs for the following UTXOs: %v", utxos)
	address := privateKey.PublicKey().Address()

	for _, utxo := range utxos {
		tx, err := CreateSingleUTXOTx(utxo, utxoAmount, utxoAmount-txFee, address, privateKey, codec)
		if err != nil {
			return nil, nil, err
		}

		txBytes = append(txBytes, tx.Bytes())
		txIDs = append(txIDs, tx.ID())
	}

	return txBytes, txIDs, nil
}
