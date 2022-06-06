package db

import (
	"context"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// write locks to see transactions more clear

	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	//run n concurrent transfer transaction

	for i := 0; i < n; i++ {

		go func() {
			ctx := context.Background()
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result

		}()
	}

	//check results

	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		//check transfer

		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		//check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		//TDD and prevent locking
		//check accounts

		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		//TODO check accounts' balance
		//fromaccount已转，减少了，toaccount已收，增加了
		//account1.balance 是一直没变的最原始的， account2.balance 也是一直没有变的最原始的
		//fromAccount.balance 是一直在减少的 toaccountbalance是一直在增加的， 每转一次增加/减少一次

		//also a lock

		diff1 := account1.Balance - fromAccount.Balance //should be equal to the money out of account1
		diff2 := toAccount.Balance - account2.Balance   // should be money into account2
		require.Equal(t, diff1, diff2)                  // if works correclty, these 2 should be the same
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) //divisible  5times of transaction1* amount, 2*amount, 3*amount, 4*amount,... n*amount

		k := int(diff1 / amount)          //k一定是一个整数
		require.True(t, k >= 1 && k <= n) //n is times of transaction
		require.NotContains(t, existed, k)
		existed[k] = true

	}

	//finally check the final updated balances
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	//account1 和 account2 的balance是一直不变的最原始的， updatedAccount1.balance是account1 所有转账的集合， updatedaccount2.balance 是擦count所有转账的集合

	//the same lock, print out after transaction

	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)

}
