package db

import (
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
	"golang.org/x/crypto/bcrypt"
)

type userRepository interface {
	Login(username string, password string) (bool, error)
	Register(username string, password string) (bool, error)
}

type UserRepository struct {
	userRepository
	db *pgx.ConnPool
}

func (r *UserRepository) Login(username string, password string) (bool, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return false, err
	}

	conn, err := r.db.Acquire()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	query, args, err := squirrel.
		Select("id").
		From("user.users").
		Where(squirrel.Eq{"username": username, "password": hashedPassword}).
		Limit(1).
		ToSql()

	id := 0
	row := conn.QueryRow(query, args...)
	err = row.Scan(&id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *UserRepository) Register(username string, password string) (bool, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return false, err
	}

	query, args, err := squirrel.
		Insert("user.users").
		Columns("username, password").
		Values(username, string(hashedPassword)).
		ToSql()
	if err != nil {
		return false, err
	}

	conn, err := r.db.Acquire()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	_, err = conn.Exec(query, args...)
	if err != nil {
		return false, err
	}

	return true, nil
}
