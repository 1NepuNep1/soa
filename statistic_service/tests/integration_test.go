package test

import (
	"context"
	"statisticservice/database"
	"statisticservice/stats"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.Conn = db
	return mock, func() { db.Close() }
}

func TestGetPostStats(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	query := "SELECT sumMerge\\(cnt\\) FROM post_events WHERE post_id = \\? AND event_type = \\?"

	mock.ExpectQuery(query).WithArgs(101, int8(1)).WillReturnRows(sqlmock.NewRows([]string{"sumMerge(cnt)"}).AddRow(5))
	mock.ExpectQuery(query).WithArgs(101, int8(2)).WillReturnRows(sqlmock.NewRows([]string{"sumMerge(cnt)"}).AddRow(3))
	mock.ExpectQuery(query).WithArgs(101, int8(3)).WillReturnRows(sqlmock.NewRows([]string{"sumMerge(cnt)"}).AddRow(2))

	metrics, err := stats.GetPostStats(context.Background(), 101)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), metrics.Views)
	assert.Equal(t, uint64(3), metrics.Likes)
	assert.Equal(t, uint64(2), metrics.Comments)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPostDynamics(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	query := `(?i)^SELECT toString\(date\) AS date, sumMerge\(cnt\) AS count FROM post_events WHERE post_id = \? AND event_type = \? GROUP BY date ORDER BY date$`

	mock.ExpectQuery(query).
		WithArgs(101, int8(1)).
		WillReturnRows(sqlmock.NewRows([]string{"date", "count"}).
			AddRow("2025-06-13", 1).
			AddRow("2025-06-14", 2),
		)

	data, err := stats.GetPostDynamics(context.Background(), 101, 1)
	assert.NoError(t, err)
	assert.Len(t, data, 2)
	assert.Equal(t, "2025-06-13", data[0].Date)
	assert.Equal(t, uint64(1), data[0].Count)
	assert.Equal(t, "2025-06-14", data[1].Date)
	assert.Equal(t, uint64(2), data[1].Count)
}

func TestGetTopPosts(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	query := `(?i)^SELECT post_id, sumMerge\(cnt\) AS count FROM post_events WHERE event_type = \? GROUP BY post_id ORDER BY count DESC LIMIT 5$`

	mock.ExpectQuery(query).
		WithArgs(int8(2)).
		WillReturnRows(sqlmock.NewRows([]string{"post_id", "count"}).
			AddRow(42, 5).
			AddRow(43, 4),
		)

	data, err := stats.GetTopPosts(context.Background(), 2, 5)
	assert.NoError(t, err)
	assert.Len(t, data, 2)
	assert.Equal(t, uint32(42), data[0].PostID)
	assert.Equal(t, uint64(5), data[0].Count)
}

func TestGetTopUsers(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	query := `(?i)^SELECT user_id, sumMerge\(cnt\) AS count FROM post_events WHERE event_type = \? GROUP BY user_id ORDER BY count DESC LIMIT 5$`

	mock.ExpectQuery(query).
		WithArgs(int8(3)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "count"}).
			AddRow(1, 10).
			AddRow(2, 7),
		)

	data, err := stats.GetTopUsers(context.Background(), 3, 5)
	assert.NoError(t, err)
	assert.Len(t, data, 2)
	assert.Equal(t, uint32(1), data[0].UserID)
	assert.Equal(t, uint64(10), data[0].Count)
}
