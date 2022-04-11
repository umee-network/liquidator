package liquidator

// TODO: We can test individual base functions for input/output,
// and even run the liquidator for a few seconds before canceling
// after replacing client-requiring steps with mockups using
// Customize() to test it as a whole.

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"github.com/umee-network/liquidator/types"
)

const (
	atomDenom = "ibc/0000000000000000000000000000000000000000000000000000000000000000"
	umeeDenom = "uumee"
)

var configBytes = []byte(`
	[liquidator]
	wait="1s"
	
	[liquidator.select]
	reward_denoms=[
	  "uumee",
	  "ibc/0000000000000000000000000000000000000000000000000000000000000000"
	]
	repay_denoms=[
	  "uumee"
	]
	`)

type IntegrationTestSuite struct {
	suite.Suite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) TestConfig() {
	// toml config file
	config := koanf.New(".")
	config.Load(rawbytes.Provider(configBytes), toml.Parser())

	s.Require().Equal(time.Second, config.Duration(ConfigKeyWait))
	s.Require().Equal([]string{umeeDenom, atomDenom}, config.Strings(ConfigKeySelectRewardDenoms))
	s.Require().Equal([]string{umeeDenom}, config.Strings(ConfigKeySelectRepayDenoms))
}

func (s *IntegrationTestSuite) TestBaseSelectFunc() {
	// toml config file with reward and repay denom preferences
	config := koanf.New(".")
	config.Load(rawbytes.Provider(configBytes), toml.Parser())

	type selectTestCase struct {
		Borrowed    []string
		Collateral  []string
		Ok          bool
		RepayDenom  string
		RewardDenom string
	}

	testCases := []selectTestCase{
		{
			// no borrowed denom
			[]string{},
			[]string{umeeDenom},
			false,
			"",
			"",
		},
		{
			// no collateral denom
			[]string{umeeDenom},
			[]string{},
			false,
			"",
			"",
		},
		{
			// only unwanted borrowed denoms
			[]string{"abcd"},
			[]string{umeeDenom},
			false,
			"",
			"",
		},
		{
			// only unwanted collateral denoms
			[]string{umeeDenom},
			[]string{"abcd"},
			false,
			"",
			"",
		},
		{
			// simple case
			[]string{umeeDenom},
			[]string{atomDenom},
			true,
			umeeDenom,
			atomDenom,
		},
		{
			// mixed in unwanted borrow
			[]string{umeeDenom, "abcd"},
			[]string{atomDenom},
			true,
			umeeDenom,
			atomDenom,
		},
		{
			// mixed in unwanted collateral
			[]string{umeeDenom},
			[]string{"abcd", atomDenom},
			true,
			umeeDenom,
			atomDenom,
		},
		{
			// config prefers umee reward over atom
			[]string{umeeDenom},
			[]string{umeeDenom, atomDenom},
			true,
			umeeDenom,
			umeeDenom,
		},
		{
			// config prefers umee reward over atom - order switched
			[]string{umeeDenom},
			[]string{atomDenom, umeeDenom},
			true,
			umeeDenom,
			umeeDenom,
		},
	}

	for _, tc := range testCases {
		borrowed := sdk.NewCoins()
		for _, denom := range tc.Borrowed {
			borrowed = borrowed.Add(sdk.NewCoin(denom, sdk.OneInt()))
		}
		collateral := sdk.NewCoins()
		for _, denom := range tc.Collateral {
			collateral = collateral.Add(sdk.NewCoin(denom, sdk.OneInt()))
		}
		target := types.LiquidationTarget{
			Addr:       []byte("addr0"),
			Borrowed:   borrowed,
			Collateral: collateral,
		}
		order, ok, err := baseSelectFunc(context.Background(), config, target)
		s.Require().NoError(err)
		s.Require().Equal(tc.Ok, ok)
		if tc.Ok {
			expected := types.LiquidationOrder{
				Addr:   []byte("addr0"),
				Repay:  sdk.NewCoin(tc.RepayDenom, sdk.OneInt()),
				Reward: sdk.NewCoin(tc.RewardDenom, sdk.ZeroInt()),
			}
			s.Require().Equal(expected, order)
		}
	}
}

func (s *IntegrationTestSuite) TestBaseApproveFunc() {
	type approveTestCase struct {
		Addr         string
		RepayDenom   string
		RepayAmount  int64
		RewardDenom  string
		RewardAmount int64
		Ok           bool
		Error        string
	}

	testCases := []approveTestCase{
		{
			"",
			"uumee",
			1,
			"uatom",
			1,
			false,
			"empty address",
		},
		{
			"addr",
			"",
			1,
			"uatom",
			1,
			false,
			"invalid denom: ",
		},
		{
			"addr",
			"uumee",
			1,
			"",
			1,
			false,
			"invalid denom: ",
		},
		{
			// reject zero reward
			"addr",
			"uumee",
			1,
			"uatom",
			0,
			false,
			"",
		},
		{
			// accept positive reward
			"addr",
			"uumee",
			1,
			"uatom",
			1,
			true,
			"",
		},
	}

	for _, tc := range testCases {
		order := types.LiquidationOrder{
			Addr: []byte(tc.Addr),
			Repay: sdk.Coin{
				Denom:  tc.RepayDenom,
				Amount: sdk.NewInt(tc.RepayAmount),
			},
			Reward: sdk.Coin{
				Denom:  tc.RewardDenom,
				Amount: sdk.NewInt(tc.RewardAmount),
			},
		}
		ok, err := baseApproveFunc(context.Background(), nil, order)
		s.Require().Equal(tc.Ok, ok)
		if tc.Error == "" {
			s.Require().NoError(err)
		} else {
			s.Require().EqualError(err, tc.Error)
		}
	}
}

func (s *IntegrationTestSuite) TestSweepLiquidation() {
	// test config: liquidator waits 1 second between loops
	config := koanf.New(".")
	config.Load(rawbytes.Provider(configBytes), toml.Parser())
	Reconfigure(config)

	// this string slice will be used to check the order liquidation steps are executed
	steps := []string{}

	// create a custom "query liquidation targets" func with empty output
	var emptyTargetFunc types.TargetFunc = func(_ context.Context, _ *koanf.Koanf,
	) ([]types.LiquidationTarget, error) {
		// log function execution
		steps = append(steps, "empty query targets")
		return nil, nil
	}

	// replace some liquidator base functions with test ones
	Customize(&emptyTargetFunc, nil, nil, nil, nil, nil)

	// start the liquidator loop
	nolog := zerolog.Nop()
	go Start(context.Background(), &nolog, "")
	time.Sleep(20 * time.Millisecond)
	defer Cancel()

	// wait for the liquidator to query for targets 3 times
	s.Require().Eventually(func() bool { return len(steps) == 3 }, 5*time.Second, 100*time.Millisecond)
	s.Require().Equal(steps, []string{
		"empty query targets",
		"empty query targets",
		"empty query targets",
	})

	// reset steps log
	steps = []string{}

	// create a custom "query liquidation targets" func with interesting predetermined output
	var customTargetFunc types.TargetFunc = func(_ context.Context, _ *koanf.Koanf,
	) ([]types.LiquidationTarget, error) {
		// log function execution
		steps = append(steps, "custom query targets")

		targets := []types.LiquidationTarget{
			{
				// addr0 will never have eligible denoms
				Addr:       sdk.AccAddress([]byte("addr0")),
				Borrowed:   sdk.NewCoins(),
				Collateral: sdk.NewCoins(),
			},
			{
				// addr1 has eligible repay and reward denoms
				Addr: sdk.AccAddress([]byte("addr1")),
				Borrowed: sdk.NewCoins(
					sdk.NewCoin(umeeDenom, sdk.NewInt(10000)),
				),
				Collateral: sdk.NewCoins(
					sdk.NewCoin(atomDenom, sdk.NewInt(200)),
				),
			},
			{
				// addr2 has repay and reward denoms, but default config refuses to repay uatom or "uabcd"
				Addr: sdk.AccAddress([]byte("addr2")),
				Borrowed: sdk.NewCoins(
					sdk.NewCoin(atomDenom, sdk.NewInt(300)),
					sdk.NewCoin("uabcd", sdk.NewInt(5)),
				),
				Collateral: sdk.NewCoins(
					sdk.NewCoin(umeeDenom, sdk.NewInt(20000)),
				),
			},
			{
				// addr3 has repay and reward denoms, of which default config prefers uumee for both
				Addr: sdk.AccAddress([]byte("addr3")),
				Borrowed: sdk.NewCoins(
					sdk.NewCoin(atomDenom, sdk.NewInt(5)),
					sdk.NewCoin(umeeDenom, sdk.NewInt(5)),
				),
				Collateral: sdk.NewCoins(
					sdk.NewCoin(atomDenom, sdk.NewInt(5)),
					sdk.NewCoin(umeeDenom, sdk.NewInt(5)),
					sdk.NewCoin("uabcd", sdk.NewInt(5)),
				),
			},
			{
				// addr4 has eligible repay and reward denoms
				Addr: sdk.AccAddress([]byte("addr4")),
				Borrowed: sdk.NewCoins(
					sdk.NewCoin(umeeDenom, sdk.NewInt(10000)),
				),
				Collateral: sdk.NewCoins(
					sdk.NewCoin(atomDenom, sdk.NewInt(200)),
				),
			},
			{
				// addr5 will never have eligible denoms
				Addr:       sdk.AccAddress([]byte("addr5")),
				Borrowed:   sdk.NewCoins(),
				Collateral: sdk.NewCoins(),
			},
		}
		return targets, nil
	}

	// wrap base select function to detect when it executes
	var customSelectFunc types.SelectFunc = func(ctx context.Context, k *koanf.Koanf, t types.LiquidationTarget,
	) (types.LiquidationOrder, bool, error) {
		// log function execution and target address
		steps = append(steps, "select from "+string(t.Addr))

		// wrap base select func, which chooses denominations based on order or priority in config file
		order, ok, err := baseSelectFunc(ctx, k, t)
		s.Require().NoError(err)
		return order, ok, err
	}

	// create a custom "estimate liquidation outcome" func
	var customEstimateFunc types.EstimateFunc = func(_ context.Context, _ *koanf.Koanf, t types.LiquidationOrder,
	) (types.LiquidationOrder, error) {
		// log function execution and target address
		steps = append(steps, fmt.Sprintf("estimate %s %s %s", string(t.Addr), t.Repay.Denom, t.Reward.Denom))

		// does not simulate transaction, just returns the input with a nonzero reward amount
		estimate := types.LiquidationOrder{
			Addr:   t.Addr,
			Repay:  t.Repay,
			Reward: t.Reward,
		}
		estimate.Reward.Amount = sdk.NewInt(1)
		return estimate, nil
	}

	// create a custom approve function which refuses to liquidate addr4
	var customApproveFunc types.ApproveFunc = func(ctx context.Context, k *koanf.Koanf, t types.LiquidationOrder,
	) (bool, error) {
		// log function execution and target address
		steps = append(steps, "approve "+string(t.Addr))
		if string(t.Addr) == "addr4" {
			return false, nil
		}

		// wrap base approve func, which says yes to any nonzero reward amount
		approved, err := baseApproveFunc(ctx, k, t)
		s.Require().NoError(err)
		return approved, err
	}

	// create a no-op "execute liquidation" func
	var customExecuteFunc types.ExecuteFunc = func(_ context.Context, _ *koanf.Koanf, t types.LiquidationOrder,
	) (types.LiquidationOrder, error) {
		// log function execution and target address
		steps = append(steps, "liquidate "+string(t.Addr))

		return types.LiquidationOrder{}, nil
	}

	// swap liquidator base functions again
	Customize(
		&customTargetFunc,
		&customSelectFunc,
		&customEstimateFunc,
		&customApproveFunc,
		&customExecuteFunc,
		nil,
	)

	// expect these exact steps to have been executed in the next tick
	s.Require().Eventually(func() bool { return len(steps) >= 15 }, 2*time.Second, 100*time.Millisecond)
	s.Require().Equal(steps, []string{
		// queries for targets once, receives addrs 0-5
		"custom query targets",
		// address with no borrow / collateral denoms
		"select from addr0",
		// address with good borrow / collateral denoms
		"select from addr1",
		fmt.Sprintf("estimate addr1 %s %s", umeeDenom, atomDenom),
		"approve addr1",
		"liquidate addr1",
		// address with valid borrow / collateral denoms, but selectFunc rejects reward denom
		"select from addr2",
		"select from addr3",
		// address with multiple borrow / collateral denoms, so selectFunc has to choose
		fmt.Sprintf("estimate addr3 %s %s", umeeDenom, umeeDenom),
		"approve addr3",
		"liquidate addr3",
		// address which approveFunc has been programmed to reject
		"select from addr4",
		fmt.Sprintf("estimate addr4 %s %s", umeeDenom, atomDenom),
		"approve addr4",
		// address with no borrow / collateral denoms
		"select from addr5",
	})
}
