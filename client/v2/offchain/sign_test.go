package offchain

import (
	"testing"


	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/x/auth/tx"
	txmodule "cosmossdk.io/x/auth/tx/config"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing" // TODO: needed as textual is not enabled by default
)

func getCodec() codec.Codec {
	registry := testutil.CodecOptions{}.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)

	return codec.NewProtoCodec(registry)
}

func MakeTestTxConfig() client.TxConfig {
	enabledSignModes := []signingtypes.SignMode{
		signingtypes.SignMode_SIGN_MODE_DIRECT,
		signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
		signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		signingtypes.SignMode_SIGN_MODE_TEXTUAL,
	}
	initClientCtx := client.Context{}
	txConfigOpts := tx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: txmodule.NewGRPCCoinMetadataQueryFn(initClientCtx),
	}
	ir, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          address.NewBech32Codec("cosmos"),
			ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
		},
	})
	if err != nil {
		panic(err)
	}
	cryptocodec.RegisterInterfaces(ir)
	cdc := codec.NewProtoCodec(ir)
	txConfig, err := tx.NewTxConfigWithOptions(cdc, txConfigOpts)
	if err != nil {
		panic(err)
	}
	return txConfig
}

func Test_getSignMode(t *testing.T) {
	tests := []struct {
		name        string
		signModeStr string
		want        apitxsigning.SignMode
	}{
		{
			name:        "direct",
			signModeStr: flags.SignModeDirect,
			want:        apitxsigning.SignMode_SIGN_MODE_DIRECT,
		},
		{
			name:        "legacy Amino JSON",
			signModeStr: flags.SignModeLegacyAminoJSON,
			want:        apitxsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		},
		{
			name:        "direct Aux",
			signModeStr: flags.SignModeDirectAux,
			want:        apitxsigning.SignMode_SIGN_MODE_DIRECT_AUX,
		},
		{
			name:        "textual",
			signModeStr: flags.SignModeTextual,
			want:        apitxsigning.SignMode_SIGN_MODE_TEXTUAL,
		},
		{
			name:        "unspecified",
			signModeStr: "",
			want:        apitxsigning.SignMode_SIGN_MODE_UNSPECIFIED,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSignMode(tt.signModeStr)
			require.Equal(t, got, tt.want)
		})
	}
}

func Test_sign(t *testing.T) {
	k := keyring.NewInMemory(getCodec())
	type args struct {
		ctx      client.Context
		fromName string
		digest   string
		signMode apitxsigning.SignMode
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "signMode direct",
			args: args{
				ctx: client.Context{
					Keyring:      k,
					TxConfig:     MakeTestTxConfig(),
					AddressCodec: address.NewBech32Codec("cosmos"),
				},
				fromName: "direct",
				digest:   "Hello world!",
				signMode: apitxsigning.SignMode_SIGN_MODE_DIRECT,
			},
		},
		{
			name: "signMode textual",
			args: args{
				ctx: client.Context{
					Keyring:      k,
					TxConfig:     MakeTestTxConfig(),
					AddressCodec: address.NewBech32Codec("cosmos"),
				},
				fromName: "textual",
				digest:   "Hello world!",
				signMode: apitxsigning.SignMode_SIGN_MODE_TEXTUAL,
			},
		},
		{
			name: "signMode LegacyAmino",
			args: args{
				ctx: client.Context{
					Keyring:      k,
					TxConfig:     MakeTestTxConfig(),
					AddressCodec: address.NewBech32Codec("cosmos"),
				},
				fromName: "legacy",
				digest:   "Hello world!",
				signMode: apitxsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := k.NewAccount(tt.args.fromName, mnemonic, tt.name, "m/44'/118'/0'/0/0", hd.Secp256k1)
			require.NoError(t, err)

			got, err := sign(tt.args.ctx, tt.args.fromName, tt.args.digest, tt.args.signMode)
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
