package crisis_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/okex/exchain/libs/tendermint/abci/types"
	"github.com/okex/exchain/libs/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/okex/exchain/libs/cosmos-sdk/simapp"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/libs/cosmos-sdk/x/crisis"
	distr "github.com/okex/exchain/libs/cosmos-sdk/x/distribution"
	"github.com/okex/exchain/libs/cosmos-sdk/x/staking"
	"github.com/okex/exchain/libs/cosmos-sdk/x/supply"
)

var (
	testModuleName        = "dummy"
	dummyRouteWhichPasses = crisis.NewInvarRoute(testModuleName, "which-passes", func(_ sdk.Context) (string, bool) { return "", false })
	dummyRouteWhichFails  = crisis.NewInvarRoute(testModuleName, "which-fails", func(_ sdk.Context) (string, bool) { return "whoops", true })
)

func createTestApp() (*simapp.SimApp, sdk.Context, []sdk.AccAddress) {
	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, 1)
	ctx := app.NewContext(true, abci.Header{})

	constantFee := sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)
	app.CrisisKeeper.SetConstantFee(ctx, constantFee)
	app.StakingKeeper.SetParams(ctx, staking.DefaultParams())

	app.CrisisKeeper.RegisterRoute(testModuleName, dummyRouteWhichPasses.Route, dummyRouteWhichPasses.Invar)
	app.CrisisKeeper.RegisterRoute(testModuleName, dummyRouteWhichFails.Route, dummyRouteWhichFails.Invar)

	feePool := distr.InitialFeePool()
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(sdk.NewCoins(constantFee)...)
	app.DistrKeeper.SetFeePool(ctx, feePool)
	app.SupplyKeeper.SetSupply(ctx, supply.NewSupply(sdk.Coins{}))

	addrs := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(10000))

	return app, ctx, addrs
}

//____________________________________________________________________________

func TestHandleMsgVerifyInvariant(t *testing.T) {
	app, ctx, addrs := createTestApp()
	sender := addrs[0]

	cases := []struct {
		name           string
		msg            sdk.Msg
		expectedResult string
	}{
		{"bad invariant route", crisis.NewMsgVerifyInvariant(sender, testModuleName, "route-that-doesnt-exist"), "fail"},
		{"invariant broken", crisis.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichFails.Route), "fail"},
		{"invariant passing", crisis.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichPasses.Route), "fail"},
		{"invalid msg", sdk.NewTestMsg(), "fail"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := crisis.NewHandler(app.CrisisKeeper)

			switch tc.expectedResult {
			case "fail":
				res, err := h(ctx, tc.msg)
				require.Error(t, err)
				require.Nil(t, res)

			case "pass":
				res, err := h(ctx, tc.msg)
				require.NoError(t, err)
				require.NotNil(t, res)

			case "panic":
				require.Panics(t, func() {
					h(ctx, tc.msg)
				})
			}
		})
	}
}

func TestHandleMsgVerifyInvariantWithNotEnoughSenderCoins(t *testing.T) {
	app, ctx, addrs := createTestApp()
	sender := addrs[0]
	coin := app.AccountKeeper.GetAccount(ctx, sender).GetCoins()[0]
	excessCoins := sdk.NewCoin(coin.Denom, coin.Amount.Add(sdk.NewDec(1)))
	app.CrisisKeeper.SetConstantFee(ctx, excessCoins)

	h := crisis.NewHandler(app.CrisisKeeper)
	msg := crisis.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichPasses.Route)

	res, err := h(ctx, msg)
	require.Error(t, err)
	require.Nil(t, res)
}

func TestHandleMsgVerifyInvariantWithInvariantBrokenAndNotEnoughPoolCoins(t *testing.T) {
	app, ctx, addrs := createTestApp()
	sender := addrs[0]

	// set the community pool to empty
	feePool := app.DistrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.DecCoins{}
	app.DistrKeeper.SetFeePool(ctx, feePool)

	h := crisis.NewHandler(app.CrisisKeeper)
	msg := crisis.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichFails.Route)

	var res *sdk.Result
	require.NotPanics(t, func() {
		res, _ = h(ctx, msg)
	}, fmt.Sprintf("%v", res))
}
