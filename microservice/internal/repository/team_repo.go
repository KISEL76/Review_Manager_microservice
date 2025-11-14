package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepo interface {
	// создаем команду
	Create(ctx context.Context, teamName string) (Team, error)

	// достаем команду по названию
	GetByName(ctx context.Context, teamName string) (Team, error)
}

type PgTeamRepo struct {
	db *pgxpool.Pool
}

func NewPgTeamRepo(db *pgxpool.Pool) *PgTeamRepo {
	return &PgTeamRepo{db: db}
}

func (r *PgTeamRepo) Create(ctx context.Context, teamName string) (Team, error) {
	const q = `INSERT INTO teams (team_name) VALUES ($1) RETURNING id, team_name`

	var t Team
	err := r.db.QueryRow(ctx, q, teamName).Scan(&t.ID, &t.Name)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == uniqueViolationErr {
			return Team{}, ErrTeamExists
		}
		return Team{}, err
	}
	return t, nil
}

func (r *PgTeamRepo) GetByName(ctx context.Context, teamName string) (Team, error) {
	const q = `SELECT id, team_name FROM teams WHERE name = "$1"`

	var t Team
	err := r.db.QueryRow(ctx, q, teamName).Scan(&t.ID, &t.Name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Team{}, ErrTeamNotFound
		}
		return Team{}, err
	}
	return t, nil
}
