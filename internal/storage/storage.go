package storage

import (
	"context"

	"github.com/Yury132/Golang-Task-1/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	// Проверка на существование пользователя
	CheckUser(ctx context.Context, email string) (bool, error)
}

type storage struct {
	conn *pgxpool.Pool
}

func (s *storage) GetUsers(ctx context.Context) ([]models.User, error) {
	query := "SELECT id, name, email FROM public.service_user"

	rows, err := s.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users = make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err = rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return users, nil
}

// Проверка на существование пользователя
func (s *storage) CheckUser(ctx context.Context, email string) (bool, error) {
	query := "SELECT id, name, email FROM public.service_user WHERE email=$1"
	//db.QueryRow("select * from Products where id = $1", 2)
	rows, err := s.conn.Query(ctx, query, email)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	// Считаем количество строк
	countRows := 0
	for rows.Next() {
		countRows++
	}
	// Если что то нашли, значит пользователь есть в БД
	check := false
	if countRows > 0 {
		check = true
	}

	if rows.Err() != nil {
		return false, err
	}

	return check, nil
}

func New(conn *pgxpool.Pool) Storage {
	return &storage{
		conn: conn,
	}
}
