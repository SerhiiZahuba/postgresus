package sqlquery

import (
	"fmt"
	"strings"
	"testing"

	"postgresus-backend/internal/features/databases"
	"postgresus-backend/internal/features/notifiers"
	"postgresus-backend/internal/features/storages"
	"postgresus-backend/internal/features/users"
	users_models "postgresus-backend/internal/features/users/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)



func getTestUserModel(t *testing.T) *users_models.User {
	t.Helper()

	signInResp := users.GetTestUser()

	us := users.GetUserService()
	u, err := us.GetFirstUser()
	if err != nil {
		t.Fatalf("failed to get first user: %v", err)
	}
	if u.ID != signInResp.UserID {
		t.Fatalf("user id mismatch: got %s, want %s", u.ID, signInResp.UserID)
	}
	return u
}


func tmpTableName() string {
	return "tmp_sqlquery_" + strings.ReplaceAll(uuid.New().String(), "-", "_")
}


func runSQL(t *testing.T, svc *Service, dbc *databases.Database, sql string) *ExecuteResponse {
	t.Helper()
	req := &ExecuteRequest{

		DatabaseID: dbc.ID,
		SQL:        sql,

	}
	resp, err := svc.Execute(dbc, req)
	assert.NoError(t, err, "execute failed for SQL: %s", sql)
	return resp
}

// ==== tests ====

func Test_Service_Execute_SelectUpdateDelete(t *testing.T) {

	testUser := getTestUserModel(t)

	testUserResp := users.GetTestUser()
	st := storages.CreateTestStorage(testUserResp.UserID)
	nt := notifiers.CreateTestNotifier(testUserResp.UserID)
	db := databases.CreateTestDatabase(testUserResp.UserID, st, nt)

	t.Cleanup(func() {
		databases.RemoveTestDatabase(db)
		notifiers.RemoveTestNotifier(nt)
		storages.RemoveTestStorage(st.ID)
	})

	dbSvc := databases.GetDatabaseService()
	dbc, err := dbSvc.GetDatabase(testUser, db.ID)
	assert.NoError(t, err)
	assert.NotNil(t, dbc)

	svc := GetSqlQueryService()

	table := tmpTableName()

	// 1) CREATE TABLE
	createSQL := fmt.Sprintf(`CREATE TABLE %s (
		id   INT PRIMARY KEY,
		name TEXT,
		cnt  INT
	)`, table)
	_ = runSQL(t, svc, dbc, createSQL)

	// 2) INSERT 3 rows
	insertSQL := fmt.Sprintf(`INSERT INTO %s (id, name, cnt) VALUES
		(1,'a',10),(2,'b',20),(3,'c',30)`, table)
	ins := runSQL(t, svc, dbc, insertSQL)
	assert.Equal(t, 3, ins.RowCount, "rows affected for INSERT should be 3")
	assert.Equal(t, 0, len(ins.Columns))
	assert.Equal(t, 0, len(ins.Rows))

	// 3) SELECT
	selectSQL := fmt.Sprintf(`SELECT id, name, cnt FROM %s ORDER BY id`, table)
	sel := runSQL(t, svc, dbc, selectSQL)
	assert.False(t, sel.Truncated)
	assert.Equal(t, 3, sel.RowCount)
	assert.Equal(t, []string{"id", "name", "cnt"}, sel.Columns)
	assert.Len(t, sel.Rows, 3)

	// simple check
	row1 := sel.Rows[0]
	assert.Equal(t, "1", fmt.Sprint(row1[0]))
	assert.Equal(t, "a", fmt.Sprint(row1[1]))
	assert.Equal(t, "10", fmt.Sprint(row1[2]))

	// 4) UPDATE two rows
	updateSQL := fmt.Sprintf(`UPDATE %s SET cnt = cnt + 5 WHERE id IN (1,3)`, table)
	upd := runSQL(t, svc, dbc, updateSQL)
	assert.Equal(t, 2, upd.RowCount)
	assert.Empty(t, upd.Columns)
	assert.Empty(t, upd.Rows)

	// 5) DELETE one rows
	deleteSQL := fmt.Sprintf(`DELETE FROM %s WHERE id = 2`, table)
	del := runSQL(t, svc, dbc, deleteSQL)
	assert.Equal(t, 1, del.RowCount)
	assert.Empty(t, del.Columns)
	assert.Empty(t, del.Rows)

	// 6) second SELECT
	sel2 := runSQL(t, svc, dbc, selectSQL)
	assert.Equal(t, 2, sel2.RowCount)
	assert.Equal(t, []string{"id", "name", "cnt"}, sel2.Columns)
	assert.Len(t, sel2.Rows, 2)

	// row with id=1
	r0 := sel2.Rows[0]
	assert.Equal(t, "1", fmt.Sprint(r0[0]))
	assert.Equal(t, "a", fmt.Sprint(r0[1]))
	assert.Equal(t, "15", fmt.Sprint(r0[2])) // 10 + 5

	// row with id=3
	r1 := sel2.Rows[1]
	assert.Equal(t, "3", fmt.Sprint(r1[0]))
	assert.Equal(t, "c", fmt.Sprint(r1[1]))
	assert.Equal(t, "35", fmt.Sprint(r1[2])) // 30 + 5

	// 7) cleanup
	dropSQL := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, table)
	_ = runSQL(t, svc, dbc, dropSQL)
}
