package db

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestStore_TransferTx(t *testing.T) {
	store := NewSqlStore(testDB)

	toAccount := createRandomAccount(t)
	fromAccount := createRandomAccount(t)

	// test transfer transaction concurrently
	amount := int64(10)
	n := 10

	errs := make(chan error)
	results := make(chan TransferTxResult)

	log.Println(fmt.Sprintf(">> before: %d, %d", fromAccount.Balance, toAccount.Balance))

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				ToAccountID:   toAccount.ID,
				FromAccountID: fromAccount.ID,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}

	var exists = make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		result := <-results

		require.NoError(t, err)

		// validate transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.NotZero(t, transfer.ID)
		require.Equal(t, transfer.FromAccountID, fromAccount.ID)
		require.Equal(t, transfer.ToAccountID, toAccount.ID)
		require.Equal(t, transfer.Amount, amount)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// validate from Entry
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.NotZero(t, fromEntry.ID)
		require.Equal(t, fromEntry.AccountID, fromAccount.ID)
		require.Equal(t, fromEntry.Amount, -amount)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		// validate to Entry
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.NotZero(t, toEntry.ID)
		require.Equal(t, toEntry.AccountID, toAccount.ID)
		require.Equal(t, toEntry.Amount, amount)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// validate accounts balance
		sender := result.FromAccount
		require.NotEmpty(t, sender)
		require.Equal(t, sender.ID, fromAccount.ID)

		receiver := result.ToAccount
		require.NotEmpty(t, receiver)
		require.Equal(t, receiver.ID, toAccount.ID)

		fmt.Println(">> tx", result.FromAccount.Balance, result.ToAccount.Balance)
		diff1 := fromAccount.Balance - sender.Balance
		diff2 := receiver.Balance - toAccount.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, exists, k)
		exists[k] = true
	}
	// check the final updated balance
	updatedSenderAccount, err := store.GetAccount(context.Background(), fromAccount.ID)
	require.NoError(t, err)

	updatedReceiverAccount, err := store.GetAccount(context.Background(), toAccount.ID)
	require.NoError(t, err)

	log.Println(fmt.Sprintf(">> after: %d, %d", updatedSenderAccount.Balance, updatedReceiverAccount.Balance))

	require.Equal(t, fromAccount.Balance-int64(n)*amount, updatedSenderAccount.Balance)
	require.Equal(t, toAccount.Balance+int64(n)*amount, updatedReceiverAccount.Balance)
}

func TestStore_TransferTxDeadlock(t *testing.T) {
	store := NewSqlStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// test transfer transaction concurrently
	amount := int64(10)
	n := 10

	errs := make(chan error)

	log.Println(fmt.Sprintf(">> before: %d, %d", account1.Balance, account2.Balance))

	for i := 0; i < n; i++ {
		go func() {
			fromAccount := account1
			toAccount := account2
			if i%2 == 1 {
				fromAccount = account2
				toAccount = account1
			}
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
			})
			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs

		require.NoError(t, err)

	}
	// check the final updated balance
	updatedSenderAccount, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedReceiverAccount, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	log.Println(fmt.Sprintf(">> after: %d, %d", updatedSenderAccount.Balance, updatedReceiverAccount.Balance))

	require.Equal(t, account1.Balance, updatedSenderAccount.Balance)
	require.Equal(t, account2.Balance, updatedReceiverAccount.Balance)
}
