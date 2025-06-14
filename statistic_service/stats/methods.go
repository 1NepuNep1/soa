package stats

import (
	"context"
	"database/sql"
	"fmt"
	"statisticservice/database"
)

type PostMetrics struct {
	Views    uint64
	Likes    uint64
	Comments uint64
}

type DayCount struct {
	Date  string
	Count uint64
}

type PostStat struct {
	PostID uint32
	Count  uint64
}

type UserStat struct {
	UserID uint32
	Count  uint64
}

func GetPostStats(ctx context.Context, postID uint32) (PostMetrics, error) {
	var metrics PostMetrics
	queries := map[int8]*uint64{
		1: &metrics.Views,
		2: &metrics.Likes,
		3: &metrics.Comments,
	}
	for evtType, dest := range queries {
		query := `SELECT sumMerge(cnt) FROM post_events WHERE post_id = ? AND event_type = ?`
		if err := database.Conn.QueryRowContext(ctx, query, postID, evtType).Scan(dest); err != nil && err != sql.ErrNoRows {
			return PostMetrics{}, err
		}
	}
	return metrics, nil
}

func GetPostDynamics(ctx context.Context, postID uint32, eventType int8) ([]DayCount, error) {
	sql := `
	SELECT toString(date) AS date, sumMerge(cnt) AS count
	FROM post_events
	WHERE post_id = ? AND event_type = ?
	GROUP BY date ORDER BY date
	`
	rows, err := database.Conn.QueryContext(ctx, sql, postID, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DayCount
	for rows.Next() {
		var dc DayCount
		if err := rows.Scan(&dc.Date, &dc.Count); err != nil {
			return nil, err
		}
		result = append(result, dc)
	}
	return result, nil
}

func GetTopPosts(ctx context.Context, eventType int8, limit int) ([]PostStat, error) {
	sql := fmt.Sprintf(`
	SELECT post_id, sumMerge(cnt) AS count
	FROM post_events
	WHERE event_type = ?
	GROUP BY post_id
	ORDER BY count DESC
	LIMIT %d
	`, limit)

	rows, err := database.Conn.QueryContext(ctx, sql, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PostStat
	for rows.Next() {
		var p PostStat
		if err := rows.Scan(&p.PostID, &p.Count); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func GetTopUsers(ctx context.Context, eventType int8, limit int) ([]UserStat, error) {
	sql := fmt.Sprintf(`
	SELECT user_id, sumMerge(cnt) AS count
	FROM post_events
	WHERE event_type = ?
	GROUP BY user_id
	ORDER BY count DESC
	LIMIT %d
	`, limit)

	rows, err := database.Conn.QueryContext(ctx, sql, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []UserStat
	for rows.Next() {
		var u UserStat
		if err := rows.Scan(&u.UserID, &u.Count); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, nil
}
