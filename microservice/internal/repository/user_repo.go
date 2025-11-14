package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo interface {
	// возвращаем пользователя по его id
	GetByID(ctx context.Context, userID string) (User, error)

	// обновляем флаг активности пользователя и возвращаем обновленную запись
	SetActive(ctx context.Context, userID string, isActive bool) (User, error)

	// возвращаем список PR в короткой форме, где пользователь назначен ревьювером
	GetReviewPrs(ctx context.Context, userID string) ([]PullRequestShort, error)

	// возвращаем здесь limit активных пользователей команды teamID,
	// исключая пользователя с excludeUserID (это для выбора нового ревьюевера нужно будет)
	FindActiveInTeamExcept(ctx context.Context, teamID int, excludeUserID string, limit int) ([]User, error)

	// вставляем или обновляем участников команды в заданной teamID
	UpsertTeamMembers(ctx context.Context, teamID int, members []User) error
}

type PgUserRepo struct {
	db *pgxpool.Pool
}

func NewPgUserRepo(db *pgxpool.Pool) *PgUserRepo {
	return &PgUserRepo{db: db}
}

func (r *PgUserRepo) GetByID(ctx context.Context, userID string) (User, error) {
	const q = `
	SELECT u.user_id,
		u.username,
		u.team_id,
		t.team_name,
		u.is_active
	FROM users u
	JOIN teams t ON t.id = u.team_id
	WHERE u.user_id = $1
	`

	var u User
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&u.ID,
		&u.Username,
		&u.TeamID,
		&u.TeamName,
		&u.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return u, nil
}

func (r *PgUserRepo) SetActive(ctx context.Context, userID string, isActive bool) (User, error) {
	const q = `
		UPDATE users u
		SET is_active = $1
		FROM teams t
		WHERE u.user_id = $2 AND u.team_id = t.id
		RETURNING u.user_id, u.username, u.team_id, t.team_name, u.is_active
	`

	var u User
	err := r.db.QueryRow(ctx, q, isActive, userID).Scan(
		&u.ID,
		&u.Username,
		&u.TeamID,
		&u.TeamName,
		&u.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return u, nil
}

func (r *PgUserRepo) GetReviewPrs(ctx context.Context, userID string) ([]PullRequestShort, error) {
	const q = `
		SELECT pr.pull_request_id,
			pr.pull_request_name,
			pr.author_id,
			s.name AS status
		FROM pull_request_reviewers r
		JOIN pull_requests pr ON pr.pull_request_id = r.pull_request_id
		JOIN pr_statuses s ON s.id = pr.status_id
		WHERE r.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []PullRequestShort
	for rows.Next() {
		var pr PullRequestShort
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.StatusName); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, rows.Err()
}

func (r *PgUserRepo) FindActiveInTeamExcept(
	ctx context.Context,
	teamID int,
	excludeUserID string,
	limit int,
) ([]User, error) {
	const q = `
		SELECT u.user_id,
			u.username,
			u.team_id,
			t.team_name,
			u.is_active
		FROM users u
		JOIN teams t ON t.id = u.team_id
		WHERE u.team_id = $1
		AND u.is_active = TRUE
		AND u.user_id <> $2
		ORDER BY random()
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, q, teamID, excludeUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamID, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	return res, rows.Err()
}

func (r *PgUserRepo) UpsertTeamMembers(ctx context.Context, teamID int, members []User) error {
	const q = `
		INSERT INTO users(user_id, username, team_id, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET username = EXCLUDED.username,
			team_id  = EXCLUDED.team_id,
			is_active = EXCLUDED.is_active
	`
	for _, m := range members {
		if _, err := r.db.Exec(ctx, q, m.ID, m.Username, teamID, m.IsActive); err != nil {
			return err
		}
	}
	return nil
}
