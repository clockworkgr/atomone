package keeper

import (
	"google.golang.org/protobuf/encoding/protowire"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	corestoretypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkgov "github.com/cosmos/cosmos-sdk/x/gov/types"
	sdkv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	v1 "github.com/atomone-hub/atomone/x/gov/types/v1"
)

// legacyParamsMarkerField is a field number that only the pre-v4 atomone.gov.v1
// Params layout defines (min_deposit_ratio). The cosmos.gov.v1 layout used since
// the v4 upgrade renumbered every field from 15 upwards by one and has no field 15,
// so its presence identifies bytes written before the migration.
const legacyParamsMarkerField = 15

// NewParamsItem returns a Params collection for the gov store that decodes both the
// current cosmos.gov.v1 layout and the atomone.gov.v1 layout written before the v4
// upgrade. The v4 upgrade migrated the stored params, but a historical query runs the
// current binary against pre-migration state; decoding those bytes with the current
// type silently shifts every field from 15 upwards (v1 and v2 state) or fails on the
// wire type of field 23 (v3 state), and the missing quorum ranges then panic in
// GetQuorum. Install it on the keeper after construction:
//
//	govKeeper.Params = keeper.NewParamsItem(cdc, govStoreService)
func NewParamsItem(cdc codec.Codec, storeService corestoretypes.KVStoreService) collections.Item[sdkv1.Params] {
	sb := collections.NewSchemaBuilder(storeService)
	item := collections.NewItem(sb, sdkgov.ParamsKey, "params", ParamsValueCodec(cdc))
	if _, err := sb.Build(); err != nil {
		panic(err)
	}
	return item
}

// ParamsValueCodec is the value codec behind NewParamsItem. Encoding always uses the
// current layout; decoding picks the layout from the bytes themselves.
func ParamsValueCodec(cdc codec.Codec) collcodec.ValueCodec[sdkv1.Params] {
	return legacyAwareParamsCodec{
		ValueCodec: codec.CollValue[sdkv1.Params](cdc),
		cdc:        cdc,
	}
}

type legacyAwareParamsCodec struct {
	collcodec.ValueCodec[sdkv1.Params]
	cdc codec.Codec
}

// Decode implements collcodec.ValueCodec.
func (c legacyAwareParamsCodec) Decode(b []byte) (sdkv1.Params, error) {
	if !IsLegacyParamsLayout(b) {
		return c.ValueCodec.Decode(b)
	}
	legacy := new(v1.Params)
	if err := c.cdc.Unmarshal(b, legacy); err != nil {
		return sdkv1.Params{}, err
	}
	params := v1.ConvertAtomOneParamsToSDK(legacy)
	fillStaticQuorumRanges(params)
	return *params, nil
}

// IsLegacyParamsLayout reports whether b is a gov Params value in the pre-v4
// atomone.gov.v1 layout. Malformed input is treated as the current layout so the
// regular codec reports the error.
func IsLegacyParamsLayout(b []byte) bool {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return false
		}
		if num == legacyParamsMarkerField {
			return true
		}
		b = b[n:]
		n = protowire.ConsumeFieldValue(num, typ, b)
		if n < 0 {
			return false
		}
		b = b[n:]
	}
	return false
}

// fillStaticQuorumRanges gives params written before the dynamic quorum existed
// (v1 and v2 state) the quorum ranges the current code expects. A range whose
// bounds both equal the static quorum makes GetQuorum return exactly the quorum
// the chain applied at that height, whatever the participation EMA.
func fillStaticQuorumRanges(params *sdkv1.Params) {
	if params.QuorumRange == nil {
		params.QuorumRange = &sdkv1.QuorumRange{Min: params.Quorum, Max: params.Quorum}
	}
	if params.ConstitutionAmendmentQuorumRange == nil {
		params.ConstitutionAmendmentQuorumRange = &sdkv1.QuorumRange{Min: params.ConstitutionAmendmentQuorum, Max: params.ConstitutionAmendmentQuorum}
	}
	if params.LawQuorumRange == nil {
		params.LawQuorumRange = &sdkv1.QuorumRange{Min: params.LawQuorum, Max: params.LawQuorum}
	}
}
