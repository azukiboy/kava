package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
	"github.com/stretchr/testify/suite"
)

type withdrawTestSuite struct {
	testutil.Suite
}

func (suite *withdrawTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
}

func TestWithdrawTestSuite(t *testing.T) {
	suite.Run(t, new(withdrawTestSuite))
}

func (suite *withdrawTestSuite) TestWithdraw_NoVaultRecord() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	withdrawAmount := sdk.NewInt64Coin(vaultDenom, 100)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	// Withdraw without having any prior deposits
	err := suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), withdrawAmount)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultRecordNotFound)

	// No balance changes
	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance),
	)

	suite.ModuleAccountBalanceEqual(
		sdk.NewCoins(),
	)
}

func (suite *withdrawTestSuite) TestWithdraw_NoVaultShareRecord() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)

	acc1DepositAmount := sdk.NewCoin(vaultDenom, sdk.NewInt(100))
	acc2WithdrawAmount := sdk.NewInt64Coin(vaultDenom, 100)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	// Create deposit from acc1 so the VaultRecord exists in state
	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0)
	err := suite.Keeper.Deposit(suite.Ctx, acc1.GetAddress(), acc1DepositAmount)
	suite.Require().NoError(err)

	acc2 := suite.CreateAccount(sdk.NewCoins(startBalance), 1)

	// Withdraw from acc2 without having any prior deposits
	err = suite.Keeper.Withdraw(suite.Ctx, acc2.GetAddress(), acc2WithdrawAmount)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultShareRecordNotFound)

	// No balance changes in acc2
	suite.AccountBalanceEqual(
		acc2.GetAddress(),
		sdk.NewCoins(startBalance),
	)

	suite.VaultTotalValuesEqual(sdk.NewCoins(acc1DepositAmount))
	suite.VaultTotalSharesEqual(types.NewVaultShares(
		types.NewVaultShare(acc1DepositAmount.Denom, acc1DepositAmount.Amount.ToDec()),
	))
}

func (suite *withdrawTestSuite) TestWithdraw_ExceedBalance() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)
	withdrawAmount := sdk.NewInt64Coin(vaultDenom, 200)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount)
	suite.Require().NoError(err)

	err = suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), withdrawAmount)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrInsufficientValue)

	// Balances still the same after deposit
	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance.Sub(depositAmount)),
	)

	suite.VaultTotalValuesEqual(sdk.NewCoins(depositAmount))
	suite.VaultTotalSharesEqual(types.NewVaultShares(
		types.NewVaultShare(depositAmount.Denom, depositAmount.Amount.ToDec()),
	))
}

func (suite *withdrawTestSuite) TestWithdraw_Zero() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	withdrawAmount := sdk.NewInt64Coin(vaultDenom, 0)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), withdrawAmount)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrInsufficientAmount)

	// No changes in balances

	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance),
	)

	suite.ModuleAccountBalanceEqual(
		sdk.NewCoins(),
	)
}

func (suite *withdrawTestSuite) TestWithdraw_InvalidVault() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	withdrawAmount := sdk.NewInt64Coin(vaultDenom, 1001)

	// Vault not created -- doesn't exist

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), withdrawAmount)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrInvalidVaultDenom)

	// No changes in balances

	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance),
	)

	suite.ModuleAccountBalanceEqual(
		sdk.NewCoins(),
	)
}

func (suite *withdrawTestSuite) TestWithdraw_FullBalance() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)
	withdrawAmount := sdk.NewInt64Coin(vaultDenom, 100)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount)
	suite.Require().NoError(err)

	err = suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), withdrawAmount)
	suite.Require().NoError(err)

	// No net changes in balances
	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance),
	)

	suite.ModuleAccountBalanceEqual(
		sdk.NewCoins(),
	)
}

func (suite *withdrawTestSuite) TestWithdraw_Partial() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)
	partialWithdrawAmount := sdk.NewInt64Coin(vaultDenom, 50)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount)
	suite.Require().NoError(err)

	err = suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), partialWithdrawAmount)
	suite.Require().NoError(err)

	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance.Sub(depositAmount).Add(partialWithdrawAmount)),
	)

	// Second withdraw for remaining 50
	err = suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), partialWithdrawAmount)
	suite.Require().NoError(err)

	// No more balance to withdraw
	err = suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), partialWithdrawAmount)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultRecordNotFound, "vault record should be deleted after no more supplied")

	// No net changes in balances
	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance),
	)

	suite.ModuleAccountBalanceEqual(
		sdk.NewCoins(),
	)
}