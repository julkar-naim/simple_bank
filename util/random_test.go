package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRandomInt(t *testing.T) {
	n := RandomInt(1000)
	require.True(t, n <= 1000)
}

func TestRandomString(t *testing.T) {
	s := RandomString(5)
	require.True(t, len(s) <= 5)
}

func TestAccount(t *testing.T) {
	owner := RandomOwner()
	money := RandomMoney()
	currency := RandomCurrency()
	require.NotEmpty(t, owner)
	require.NotZero(t, money)
	require.NotEmpty(t, currency)
	// test
}
