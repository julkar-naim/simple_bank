package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/julkar-naim/simple-bank/db/mock"
	db "github.com/julkar-naim/simple-bank/db/sqlc"
	"github.com/julkar-naim/simple-bank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAccountAPI(t *testing.T) {

	testCases := []struct {
		name          string
		RequestBody   createAccountRequest
		buildStub     func(store mockdb.MockStore, ctrl *gomock.Controller, account db.Account)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder, account db.Account)
	}{
		{
			"OK",
			randomCreateAccountData(),
			func(store mockdb.MockStore, ctrl *gomock.Controller, account db.Account) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(account, nil)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder, account db.Account) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			"BadRequest",
			createAccountRequest{
				Owner:    "",
				Currency: "xyz",
			},
			func(store mockdb.MockStore, ctrl *gomock.Controller, account db.Account) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder, account db.Account) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			"InternalServer",
			randomCreateAccountData(),
			func(store mockdb.MockStore, ctrl *gomock.Controller, account db.Account) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder, account db.Account) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			account := db.Account{
				ID:       util.RandomInt(1000),
				Owner:    tc.RequestBody.Owner,
				Balance:  0,
				Currency: tc.RequestBody.Currency,
			}

			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(*store, ctrl, account)

			// configure test server
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodPost, "/accounts", buildRequestBody(tc.RequestBody))
			require.NoError(t, err)

			// start test server
			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder, account)
		})
	}

}

func TestUpdateAccountAPI(t *testing.T) {

	account1 := randomAccount()
	account2 := randomAccount()

	updateAccountReq := updateAccountRequest{
		ID:       account1.ID,
		Owner:    account2.Owner,
		Balance:  account2.Balance,
		Currency: account2.Currency,
	}

	updatedAccount := db.Account{
		ID:       account1.ID,
		Owner:    account2.Owner,
		Balance:  account2.Balance,
		Currency: account2.Currency,
	}
	invalidRequest := updateAccountReq
	invalidRequest.ID = -1
	invalidRequest.Currency = "xyz"

	testCases := []struct {
		name          string
		RequestBody   updateAccountRequest
		buildStub     func(store mockdb.MockStore, ctrl *gomock.Controller)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			"OK",
			updateAccountReq,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().UpdateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(updatedAccount, nil)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, updatedAccount)
			},
		},
		{
			"BadRequest",
			invalidRequest,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			"InternalServerError",
			updateAccountReq,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().UpdateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(updatedAccount, sql.ErrConnDone)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(*store, ctrl)

			// configure test server
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodPost, "/accounts/update", buildRequestBody(tc.RequestBody))
			require.NoError(t, err)

			// start test server
			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder)
		})
	}

}

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		ID            int64
		buildStub     func(store mockdb.MockStore, ctrl *gomock.Controller)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			"OK",
			account.ID,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			"NotFound",
			account.ID,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			"InternalServerError",
			account.ID,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			"BadRequest",
			0,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(*store, ctrl)

			// configure test server
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.ID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// start test server
			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder)
		})
	}

}

func TestGetAccountListAPI(t *testing.T) {
	var accounts []db.Account
	n := 5

	for i := 0; i < n; i++ {
		accounts = append(accounts, randomAccount())
	}

	testCases := []struct {
		name          string
		PageID        int32
		PageSize      int32
		buildStub     func(store mockdb.MockStore, ctrl *gomock.Controller)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			"OK",
			1,
			5,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return(accounts, nil)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			"BadRequest",
			0,
			4,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			"InternalServerError",
			1,
			5,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(*store, ctrl)

			// configure test server
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts?page_id=%d&page_size=%d", tc.PageID, tc.PageSize)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// start test server
			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder)
		})
	}

}

func TestDeleteAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		ID            int64
		buildStub     func(store mockdb.MockStore, ctrl *gomock.Controller)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			"OK",
			account.ID,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			"InternalServerError",
			account.ID,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(sql.ErrConnDone)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			"BadRequest",
			0,
			func(store mockdb.MockStore, ctrl *gomock.Controller) {
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(*store, ctrl)

			// configure test server
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/delete/%d", tc.ID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// start test server
			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder)
		})
	}

}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func randomCreateAccountData() createAccountRequest {
	return createAccountRequest{
		Owner:    util.RandomOwner(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var bodyAccount db.Account
	err = json.Unmarshal(data, &bodyAccount)
	require.NoError(t, err)
	require.Equal(t, bodyAccount, account)
}

func buildRequestBody(data any) io.Reader {
	body, _ := json.Marshal(data)
	return bytes.NewReader(body)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var bodyAccount []db.Account
	err = json.Unmarshal(data, &bodyAccount)
	require.NoError(t, err)
	require.Equal(t, bodyAccount, accounts)
}
