package keeper_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/atomone-hub/atomone/x/gov/keeper"
)

// Raw gov params read from atomone-1 state through abci_query path="/store/gov/key",
// data=0x30, at heights inside each params layout era.
const (
	// height 749400: v1.0.0 state, before the dynamic quorum existed
	paramsPreV3 = "ChMKBnVhdG9uZRIJNTEyMDAwMDAwEgQIgOpJGgQIgN9uIhQwLjI1MDAwMDAwMDAwMDAwMDAwMCoUMC42NjcwMDAwMDAwMDAwMDAwMDA6FDAuMTAwMDAwMDAwMDAwMDAwMDAwcAF6FDAuMDEwMDAwMDAwMDAwMDAwMDAwggEUMC4yNTAwMDAwMDAwMDAwMDAwMDCKARQwLjkwMDAwMDAwMDAwMDAwMDAwMJIBFDAuMjUwMDAwMDAwMDAwMDAwMDAwmgEUMC45MDAwMDAwMDAwMDAwMDAwMDCiAQQIgLxpqgEECICjBQ=="
	// height 6085557: v3 state, atomone.gov.v1 layout with quorum ranges
	paramsV3 = "ChMKBnVhdG9uZRIJNTEyMDAwMDAwEgQIgOpJGgQIgN9uIhQwLjI1MDAwMDAwMDAwMDAwMDAwMCoUMC42NjcwMDAwMDAwMDAwMDAwMDA6FDAuMTAwMDAwMDAwMDAwMDAwMDAwcAF6FDAuMDEwMDAwMDAwMDAwMDAwMDAwggEUMC4yNTAwMDAwMDAwMDAwMDAwMDCKARQwLjkwMDAwMDAwMDAwMDAwMDAwMJIBFDAuMjUwMDAwMDAwMDAwMDAwMDAwmgEUMC45MDAwMDAwMDAwMDAwMDAwMDCiAQQIgLxpqgEECICjBboBSgoSCgZ1YXRvbmUSCDEwMDAwMDAwEgQIgPUkGAIiFDAuMDUwMDAwMDAwMDAwMDAwMDAwKhQwLjAyNTAwMDAwMDAwMDAwMDAwMDACwgFIChAKBnVhdG9uZRIGMTAwMDAwEgQIgKMFGAUiFDAuMDEwMDAwMDAwMDAwMDAwMDAwKhQwLjAwNTAwMDAwMDAwMDAwMDAwMDACygEUMC44MDAwMDAwMDAwMDAwMDAwMDDSASwKFDAuNTAwMDAwMDAwMDAwMDAwMDAwEhQwLjEwMDAwMDAwMDAwMDAwMDAwMNoBLAoUMC41MDAwMDAwMDAwMDAwMDAwMDASFDAuMTAwMDAwMDAwMDAwMDAwMDAw4gEsChQwLjUwMDAwMDAwMDAwMDAwMDAwMBIUMC4xMDAwMDAwMDAwMDAwMDAwMDA="
	// height 10276000: v4 state, cosmos.gov.v1 layout written by the v4 migration
	paramsV4 = "EgQIgOpJGgQIgN9uKhQwLjY2NzAwMDAwMDAwMDAwMDAwMHABggEUMC4wMTAwMDAwMDAwMDAwMDAwMDCSARQwLjkwMDAwMDAwMDAwMDAwMDAwMKIBFDAuOTAwMDAwMDAwMDAwMDAwMDAwqgEECIC8abIBBAiAowXCAUsKEwoGdWF0b25lEgk1MTIwMDAwMDASBAiA9SQYAiIUMC4wNTAwMDAwMDAwMDAwMDAwMDAqFDAuMDI1MDAwMDAwMDAwMDAwMDAwMALKAUoKEgoGdWF0b25lEgg1MTIwMDAwMBIECICjBRgFIhQwLjAxMDAwMDAwMDAwMDAwMDAwMCoUMC4wMDUwMDAwMDAwMDAwMDAwMDAwAtIBFDAuODAwMDAwMDAwMDAwMDAwMDAw2gEsChQwLjUwMDAwMDAwMDAwMDAwMDAwMBIUMC4xMDAwMDAwMDAwMDAwMDAwMDDiASwKFDAuNTAwMDAwMDAwMDAwMDAwMDAwEhQwLjEwMDAwMDAwMDAwMDAwMDAwMOoBLAoUMC41MDAwMDAwMDAwMDAwMDAwMDASFDAuMTAwMDAwMDAwMDAwMDAwMDAw8gEFCIDUkwH6AQsxMDAwMDAwMDAwMA=="
)

func fixture(t *testing.T, b64 string) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(b64)
	require.NoError(t, err)
	return b
}

func newCodec() codec.Codec {
	return codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
}

func TestIsLegacyParamsLayout(t *testing.T) {
	require.True(t, keeper.IsLegacyParamsLayout(fixture(t, paramsPreV3)))
	require.True(t, keeper.IsLegacyParamsLayout(fixture(t, paramsV3)))
	require.False(t, keeper.IsLegacyParamsLayout(fixture(t, paramsV4)))

	current, err := newCodec().Marshal(&sdkv1.Params{})
	require.NoError(t, err)
	require.False(t, keeper.IsLegacyParamsLayout(current))
	defaults := sdkv1.DefaultParams()
	current, err = newCodec().Marshal(&defaults)
	require.NoError(t, err)
	require.False(t, keeper.IsLegacyParamsLayout(current))
	require.False(t, keeper.IsLegacyParamsLayout([]byte{0xff}))
}

func TestParamsValueCodecDecodesPreV3State(t *testing.T) {
	params, err := keeper.ParamsValueCodec(newCodec()).Decode(fixture(t, paramsPreV3))
	require.NoError(t, err)

	// Values as the chain stored them at genesis; every field from 15 upwards
	// would otherwise land one field off.
	require.Equal(t, "0.250000000000000000", params.Quorum)
	require.Equal(t, "0.667000000000000000", params.Threshold)
	require.Equal(t, "0.010000000000000000", params.MinDepositRatio)
	require.Equal(t, "0.250000000000000000", params.ConstitutionAmendmentQuorum)
	require.Equal(t, "0.900000000000000000", params.ConstitutionAmendmentThreshold)
	require.Equal(t, "0.250000000000000000", params.LawQuorum)
	require.Equal(t, "0.900000000000000000", params.LawThreshold)
	require.NotNil(t, params.QuorumTimeout)
	require.Equal(t, float64(20*24*3600), params.QuorumTimeout.Seconds())
	require.NotNil(t, params.MaxVotingPeriodExtension)
	require.Equal(t, float64(24*3600), params.MaxVotingPeriodExtension.Seconds())
	require.Equal(t, uint64(0), params.QuorumCheckCount)

	// No dynamic quorum yet: the ranges pin the static quorum the chain applied
	require.Equal(t, &sdkv1.QuorumRange{Min: "0.250000000000000000", Max: "0.250000000000000000"}, params.QuorumRange)
	require.Equal(t, &sdkv1.QuorumRange{Min: "0.250000000000000000", Max: "0.250000000000000000"}, params.ConstitutionAmendmentQuorumRange)
	require.Equal(t, &sdkv1.QuorumRange{Min: "0.250000000000000000", Max: "0.250000000000000000"}, params.LawQuorumRange)
}

func TestParamsValueCodecDecodesV3State(t *testing.T) {
	cdc := newCodec()
	// The current layout cannot read these bytes at all (field 23 changed wire type)
	_, err := codec.CollValue[sdkv1.Params](cdc).Decode(fixture(t, paramsV3))
	require.Error(t, err)

	params, err := keeper.ParamsValueCodec(cdc).Decode(fixture(t, paramsV3))
	require.NoError(t, err)
	require.Equal(t, "0.250000000000000000", params.Quorum)
	require.Equal(t, "0.010000000000000000", params.MinDepositRatio)
	require.Equal(t, "0.900000000000000000", params.LawThreshold)
	require.Equal(t, "0.800000000000000000", params.BurnDepositNoThreshold)
	require.NotNil(t, params.MinDepositThrottler)
	// Ranges written by the v3 upgrade are kept as stored, not replaced
	require.Equal(t, &sdkv1.QuorumRange{Min: "0.100000000000000000", Max: "0.500000000000000000"}, params.QuorumRange)
}

func TestParamsValueCodecKeepsCurrentLayout(t *testing.T) {
	cdc := newCodec()
	want, err := codec.CollValue[sdkv1.Params](cdc).Decode(fixture(t, paramsV4))
	require.NoError(t, err)
	got, err := keeper.ParamsValueCodec(cdc).Decode(fixture(t, paramsV4))
	require.NoError(t, err)
	require.Equal(t, want, got)

	// Encoding is unchanged: what the codec writes, the plain codec reads back
	defaults := sdkv1.DefaultParams()
	encoded, err := keeper.ParamsValueCodec(cdc).Encode(defaults)
	require.NoError(t, err)
	roundTrip, err := codec.CollValue[sdkv1.Params](cdc).Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, defaults, roundTrip)
}
