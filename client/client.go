package client

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/rs/zerolog"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	tmjsonclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"

	umeeappbeta "github.com/umee-network/umee/app/beta"
	umeeparams "github.com/umee-network/umee/app/params"
)

type (
	// LeverageClient defines a structure that interfaces with the Umee node.
	// Copied from umee/price-feeder, with validator address fields removed.
	LeverageClient struct {
		Logger               zerolog.Logger
		ChainID              string
		KeyringBackend       string
		KeyringDir           string
		KeyringPass          string
		TMRPC                string
		RPCTimeout           time.Duration
		LiquidatorAddr       sdk.AccAddress
		LiquidatorAddrString string
		Encoding             umeeparams.EncodingConfig
		GasPrices            string
		GasAdjustment        float64
		GRPCEndpoint         string
		KeyringPassphrase    string
	}

	passReader struct {
		pass string
		buf  *bytes.Buffer
	}
)

func NewLeverageClient(
	logger zerolog.Logger,
	chainID string,
	keyringBackend string,
	keyringDir string,
	keyringPass string,
	tmRPC string,
	rpcTimeout time.Duration,
	liquidatorAddrString string,
	grpcEndpoint string,
	gasAdjustment float64,
) (LeverageClient, error) {
	liquidatorAddr, err := sdk.AccAddressFromBech32(liquidatorAddrString)
	if err != nil {
		return LeverageClient{}, err
	}

	return LeverageClient{
		Logger:               logger.With().Str("module", "leverage_client").Logger(),
		ChainID:              chainID,
		KeyringBackend:       keyringBackend,
		KeyringDir:           keyringDir,
		KeyringPass:          keyringPass,
		TMRPC:                tmRPC,
		RPCTimeout:           rpcTimeout,
		LiquidatorAddr:       liquidatorAddr,
		LiquidatorAddrString: liquidatorAddrString,
		Encoding:             umeeappbeta.MakeEncodingConfig(),
		GasAdjustment:        gasAdjustment,
		GRPCEndpoint:         grpcEndpoint,
	}, nil
}

func newPassReader(pass string) io.Reader {
	return &passReader{
		pass: pass,
		buf:  new(bytes.Buffer),
	}
}

func (r *passReader) Read(p []byte) (n int, err error) {
	n, err = r.buf.Read(p)
	if err == io.EOF || n == 0 {
		r.buf.WriteString(r.pass + "\n")

		n, err = r.buf.Read(p)
	}

	return n, err
}

// CreateClientContext creates an SDK client Context instance used for transaction
// generation, signing and broadcasting.
func (lc LeverageClient) CreateClientContext() (client.Context, error) {
	var keyringInput io.Reader
	if len(lc.KeyringPass) > 0 {
		keyringInput = newPassReader(lc.KeyringPass)
	} else {
		return client.Context{}, fmt.Errorf("no keyring password provided")
	}

	kr, err := keyring.New("liquidator", lc.KeyringBackend, lc.KeyringDir, keyringInput)
	if err != nil {
		return client.Context{}, err
	}

	httpClient, err := tmjsonclient.DefaultHTTPClient(lc.TMRPC)
	if err != nil {
		return client.Context{}, err
	}

	httpClient.Timeout = lc.RPCTimeout

	tmRPC, err := rpchttp.NewWithClient(lc.TMRPC, "/websocket", httpClient)
	if err != nil {
		return client.Context{}, err
	}

	keyInfo, err := kr.KeyByAddress(lc.LiquidatorAddr)
	if err != nil {
		return client.Context{}, err
	}

	clientCtx := client.Context{
		ChainID:           lc.ChainID,
		JSONCodec:         lc.Encoding.Marshaler,
		InterfaceRegistry: lc.Encoding.InterfaceRegistry,
		Output:            os.Stderr,
		BroadcastMode:     flags.BroadcastSync,
		TxConfig:          lc.Encoding.TxConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		Codec:             lc.Encoding.Marshaler,
		LegacyAmino:       lc.Encoding.Amino,
		Input:             os.Stdin,
		NodeURI:           lc.TMRPC,
		Client:            tmRPC,
		Keyring:           kr,
		FromAddress:       lc.LiquidatorAddr,
		FromName:          keyInfo.GetName(),
		From:              keyInfo.GetName(),
		OutputFormat:      "json",
		UseLedger:         false,
		Simulate:          false,
		GenerateOnly:      false,
		Offline:           false,
		SkipConfirm:       true,
	}

	return clientCtx, nil
}

// CreateTxFactory creates an SDK Factory instance used for transaction
// generation, signing and broadcasting.
func (lc LeverageClient) CreateTxFactory() (tx.Factory, error) {
	clientCtx, err := lc.CreateClientContext()
	if err != nil {
		return tx.Factory{}, err
	}

	txFactory := tx.Factory{}.
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithChainID(lc.ChainID).
		WithTxConfig(clientCtx.TxConfig).
		WithGasAdjustment(lc.GasAdjustment).
		WithGasPrices(lc.GasPrices).
		WithKeybase(clientCtx.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithSimulateAndExecute(true)

	return txFactory, nil
}
