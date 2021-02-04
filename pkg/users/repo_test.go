package users

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type userRepoTestCase struct {
	name      string
	queryMock sqlQueryMock
	funcName  string
}

type sqlQueryMock struct {
	query       string
	args        []driver.Value
	columns     []string
	values      []map[string]interface{}
	err         error
	requestType string
	mockResult  driver.Result
	mockQuery   func(mock sqlmock.Sqlmock, query string, args []driver.Value, result driver.Result, rows *sqlmock.Rows, err error)
}

var UserRepoTestCases = []userRepoTestCase{
	userRepoTestCase{
		name:     "Success user creation",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into users",
			args:        []driver.Value{"example@mail.com"},
			columns:     []string{"id", "email"},
			requestType: "insert",
			mockResult:  sqlmock.NewResult(1, 1),
			mockQuery: func(mock sqlmock.Sqlmock, query string, args []driver.Value, result driver.Result, rows *sqlmock.Rows, err error) {
				mock.
					ExpectExec(query).
					WithArgs(args...).
					WillReturnResult(result)
			},
		},
	},
	userRepoTestCase{
		name:     "Failed user creation",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into users",
			args:        []driver.Value{"example@mail.com"},
			columns:     []string{"id", "email"},
			requestType: "insert-error",
			err:         fmt.Errorf("Insert error"),
			mockQuery: func(mock sqlmock.Sqlmock, query string, args []driver.Value, result driver.Result, rows *sqlmock.Rows, err error) {
				mock.
					ExpectExec(query).
					WithArgs(args...).
					WillReturnError(err)
			},
		},
	},
}

func TestUsersRepo(t *testing.T) {
	for _, tc := range UserRepoTestCases {
		testLabel := strings.Join([]string{"Repo", "User", tc.name}, " ")
		t.Run(testLabel, func(t *testing.T) {
			realArgs := []reflect.Value{}
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("cant create mock: %s", err)
			}
			defer db.Close()

			repo := UsersService{
				db: db,
			}

			rows := sqlmock.
				NewRows(tc.queryMock.columns)

			for _, row := range tc.queryMock.values {
				if tc.queryMock.err != nil {
					rows = rows.AddRow(nil, row["email"]).RowError(row["id"].(int), tc.queryMock.err)
				} else {
					rows = rows.AddRow(row["id"], row["email"])
				}
			}

			for _, arg := range tc.queryMock.args {
				realArgs = append(realArgs, reflect.ValueOf(arg))
			}
			tc.queryMock.mockQuery(mock, tc.queryMock.query, tc.queryMock.args, tc.queryMock.mockResult, rows, tc.queryMock.err)

			var result []reflect.Value
			if len(tc.queryMock.args) > 0 {
				result = reflect.ValueOf(
					&repo,
				).MethodByName(
					tc.funcName,
				).Call(realArgs)
			} else {
				result = reflect.ValueOf(
					&repo,
				).MethodByName(
					tc.funcName,
				).Call(realArgs)
			}
			var reflectErr error
			rerr := result[1].Interface()
			if rerr != nil {
				reflectErr = rerr.(error)
			}

			if reflectErr != nil && tc.queryMock.err == nil {
				t.Errorf("unexpected err: %s", reflectErr)
				return
			}

			if tc.queryMock.err != nil {
				if reflectErr == nil {
					t.Errorf("expected error, got nil: %s", reflectErr)
					return
				}
			}
		})
	}
}

func TestNewUserService(t *testing.T) {
	db, _, _ := sqlmock.New()
	repo := NewUsersService(db)
	_, correctType := repo.(UsersService)
	if !correctType {
		t.Errorf("Wrong type of UserService")
	}
}
