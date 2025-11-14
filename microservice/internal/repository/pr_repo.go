package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PrRepo interface {
	// проверяем, существует ли PR с таким идентификатором
	Exists(ctx context.Context, prID string) (bool, error)

	// создаем новый PR в базе (без назначения ревьюверов)
	Insert(ctx context.Context, pr PullRequest) error

	// добавляем к PR указанных ревьюверов
	AddReviewers(ctx context.Context, prID string, reviewerIDs []string) error

	// возвращаем PR и список его ревьюверов по идентификатору
	GetWithReviewers(ctx context.Context, prID string) (PullRequest, []string, error)

	// идемпотентно помечает PR как MERGED и устанавливает merged_at
	SetMerged(ctx context.Context, prID string, mergedAt time.Time) (PullRequest, error)

	// заменяем одного ревьювера на другого в пределах одного PR
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
}

type PgPrRepo struct {
	db *pgxpool.Pool
}

func NewPgPrRepo(db *pgxpool.Pool) *PgPrRepo {
	return &PgPrRepo{db: db}
}

func (r *PgPrRepo) Exists(ctx context.Context, prID string) (bool, error) {
	const q = `SELECT EXISTS (SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, q, prID).Scan(&exists)
	return exists, err
}

func (r *PgPrRepo) Insert(ctx context.Context, pr PullRequest) error {
	const q = `
		INSERT INTO pull_requests(pull_request_id,
    	pull_request_name,
    	author_id,
    	status_id,
    	created_at,
    	merged_at,
    	need_more_reviewers) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, q, pr.ID,
		pr.Name,
		pr.AuthorID,
		pr.StatusID,
		pr.CreatedAt,
		pr.MergedAt,
		false,
	)
	return err
}

func (r *PgPrRepo) AddReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	if len(reviewerIDs) == 0 {
		return nil
	}

	const q = `
		INSERT INTO pull_request_reviewers(pull_request_id, reviewer_id)
		VALUES ($1, $2)
		ON CONFLICT (pull_request_id, reviewer_id) DO NOTHING
	`
	for _, rid := range reviewerIDs {
		if _, err := r.db.Exec(ctx, q, prID, rid); err != nil {
			return err
		}
	}
	return nil
}

func (r *PgPrRepo) GetWithReviewers(ctx context.Context, prID string) (PullRequest, []string, error) {
	const qPR = `
		SELECT pr.pull_request_id,
			pr.pull_request_name,
			pr.author_id,
			pr.status_id,
			s.name AS status_name,
			pr.created_at,
			pr.merged_at
		FROM pull_requests pr
		JOIN pr_statuses s ON s.id = pr.status_id
		WHERE pr.pull_request_id = $1
	`

	var pr PullRequest
	var mergedAt *time.Time

	err := r.db.QueryRow(ctx, qPR, prID).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.StatusID,
		&pr.StatusName,
		&pr.CreatedAt,
		&mergedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return PullRequest{}, nil, ErrPRNotFound
		}
		return PullRequest{}, nil, err
	}
	pr.MergedAt = mergedAt

	const qRev = `SELECT reviewer_id FROM pull_request_reviewers WHERE pull_request_id = $1`

	rows, err := r.db.Query(ctx, qRev, prID)
	if err != nil {
		return PullRequest{}, nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return PullRequest{}, nil, err
		}
		reviewers = append(reviewers, id)
	}
	if err := rows.Err(); err != nil {
		return PullRequest{}, nil, err
	}

	return pr, reviewers, nil
}

func (r *PgPrRepo) SetMerged(ctx context.Context, prID string, mergedAt time.Time) (PullRequest, error) {
	const q = `
		UPDATE pull_requests
		SET status_id = $1,
			merged_at = COALESCE(merged_at, $2)
		WHERE pull_request_id = $3
		RETURNING pull_request_id, pull_request_name, author_id, status_id, created_at, merged_at
	`
	var (
		pr         PullRequest
		mergedAtDB *time.Time
	)

	err := r.db.QueryRow(ctx, q, statusMerged, mergedAt, prID).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.StatusID,
		&pr.CreatedAt,
		&mergedAtDB,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return PullRequest{}, ErrPRNotFound
		}
		return PullRequest{}, err
	}
	pr.MergedAt = mergedAtDB

	const qStatus = `SELECT name FROM pr_statuses WHERE id = $1`
	if err := r.db.QueryRow(ctx, qStatus, pr.StatusID).Scan(&pr.StatusName); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}

func (r *PgPrRepo) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	const q = `
		UPDATE pull_request_reviewers
		SET reviewer_id = $3
		WHERE pull_request_id = $1
		AND reviewer_id = $2
	`

	ct, err := r.db.Exec(ctx, q, prID, oldReviewerID, newReviewerID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrReviewerNotSet
	}
	return nil
}
