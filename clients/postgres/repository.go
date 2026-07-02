package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

// Repository — реализация интерфейса userRepository для PostgreSQL (через pgx)
type Repository struct {
	db *pgx.Conn
}

// NewRepository — конструктор, принимает подключение к БД
func NewRepository(db *pgx.Conn) *Repository {
	return &Repository{db: db}
}

// GetUserCity — получить город пользователя по userID
func (r *Repository) GetUserCity(ctx context.Context, userID int64) (string, error) {
	var city string
	query := "SELECT city FROM users WHERE id = $1"

	err := r.db.QueryRow(ctx, query, userID).Scan(&city)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Пользователь не найден — возвращаем пустую строку без ошибки
			return "", nil
		}
		log.Printf("GetUserCity error: %v", err)
		return "", fmt.Errorf("failed to get user city: %w", err)
	}

	return city, nil
}

// CreateUser — создать нового пользователя
func (r *Repository) CreateUser(ctx context.Context, userID int64) error {
	query := "INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING"

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		log.Printf("CreateUser error: %v", err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateCity — обновить город пользователя
func (r *Repository) UpdateCity(ctx context.Context, userID int64, city string) error {
	// Сначала проверяем, существует ли пользователь
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"
	err := r.db.QueryRow(ctx, checkQuery, userID).Scan(&exists)
	if err != nil {
		log.Printf("UpdateCity check user error: %v", err)
		return fmt.Errorf("failed to check user: %w", err)
	}

	if !exists {
		// Если пользователя нет — создаём
		err := r.CreateUser(ctx, userID)
		if err != nil {
			return err
		}
	}

	// Обновляем город
	updateQuery := "UPDATE users SET city = $1 WHERE id = $2"
	_, err = r.db.Exec(ctx, updateQuery, city, userID)
	if err != nil {
		log.Printf("UpdateCity update error: %v", err)
		return fmt.Errorf("failed to update city: %w", err)
	}

	log.Printf("City updated for user %d: %s", userID, city)
	return nil
}
