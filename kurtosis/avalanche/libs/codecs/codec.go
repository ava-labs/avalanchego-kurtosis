// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package helpers

import (
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/codec/hierarchycodec"
	"github.com/ava-labs/avalanchego/codec/linearcodec"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/nftfx"
	"github.com/ava-labs/avalanchego/vms/propertyfx"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

const (
	// Codec version used before AvalancheGo 1.1.0
	pre110CodecVersion = uint16(0)

	// Current codec version
	currentCodecVersion = uint16(1)
)

func CreateXChainCodec() (codec.Manager, error) {
	codecManager := codec.NewDefaultManager()

	pre110Codec := linearcodec.NewDefault()
	errs := wrappers.Errs{}
	errs.Add(
		pre110Codec.RegisterType(&avm.BaseTx{}),
		pre110Codec.RegisterType(&avm.CreateAssetTx{}),
		pre110Codec.RegisterType(&avm.OperationTx{}),
		pre110Codec.RegisterType(&avm.ImportTx{}),
		pre110Codec.RegisterType(&avm.ExportTx{}),
		pre110Codec.RegisterType(&secp256k1fx.TransferInput{}),
		pre110Codec.RegisterType(&secp256k1fx.MintOutput{}),
		pre110Codec.RegisterType(&secp256k1fx.TransferOutput{}),
		pre110Codec.RegisterType(&secp256k1fx.MintOperation{}),
		pre110Codec.RegisterType(&secp256k1fx.Credential{}),
		pre110Codec.RegisterType(&nftfx.MintOutput{}),
		pre110Codec.RegisterType(&nftfx.TransferOutput{}),
		pre110Codec.RegisterType(&nftfx.MintOperation{}),
		pre110Codec.RegisterType(&nftfx.TransferOperation{}),
		pre110Codec.RegisterType(&nftfx.Credential{}),
		pre110Codec.RegisterType(&propertyfx.MintOutput{}),
		pre110Codec.RegisterType(&propertyfx.OwnedOutput{}),
		pre110Codec.RegisterType(&propertyfx.MintOperation{}),
		pre110Codec.RegisterType(&propertyfx.BurnOperation{}),
		pre110Codec.RegisterType(&propertyfx.Credential{}),
		codecManager.RegisterCodec(pre110CodecVersion, pre110Codec),
	)
	if errs.Errored() {
		return nil, errs.Err
	}

	currentCodec := hierarchycodec.NewDefault()
	errs.Add(
		currentCodec.RegisterType(&avm.BaseTx{}),
		currentCodec.RegisterType(&avm.CreateAssetTx{}),
		currentCodec.RegisterType(&avm.OperationTx{}),
		currentCodec.RegisterType(&avm.ImportTx{}),
		currentCodec.RegisterType(&avm.ExportTx{}),
	)
	currentCodec.NextGroup()
	errs.Add(
		currentCodec.RegisterType(&secp256k1fx.TransferInput{}),
		currentCodec.RegisterType(&secp256k1fx.MintOutput{}),
		currentCodec.RegisterType(&secp256k1fx.TransferOutput{}),
		currentCodec.RegisterType(&secp256k1fx.MintOperation{}),
		currentCodec.RegisterType(&secp256k1fx.Credential{}),
	)
	currentCodec.NextGroup()
	errs.Add(
		currentCodec.RegisterType(&nftfx.MintOutput{}),
		currentCodec.RegisterType(&nftfx.TransferOutput{}),
		currentCodec.RegisterType(&nftfx.MintOperation{}),
		currentCodec.RegisterType(&nftfx.TransferOperation{}),
		currentCodec.RegisterType(&nftfx.Credential{}),
	)
	currentCodec.NextGroup()
	errs.Add(
		currentCodec.RegisterType(&propertyfx.MintOutput{}),
		currentCodec.RegisterType(&propertyfx.OwnedOutput{}),
		currentCodec.RegisterType(&propertyfx.MintOperation{}),
		currentCodec.RegisterType(&propertyfx.BurnOperation{}),
		currentCodec.RegisterType(&propertyfx.Credential{}),
		codecManager.RegisterCodec(currentCodecVersion, currentCodec),
	)
	return codecManager, errs.Err
}
