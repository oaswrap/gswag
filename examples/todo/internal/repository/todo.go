// Package repository provides SQLite-backed persistence for todos.
package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/oaswrap/gswag/examples/todo/internal/model"
)

// ErrNotFound is returned when a todo does not exist.
var ErrNotFound = errors.New("todo not found")

// TodoRepository defines the persistence contract.
type TodoRepository interface {
	FindAll() ([]model.Todo, error)
	FindByID(id int64) (*model.Todo, error)
	Create(req model.CreateTodoRequest) (*model.Todo, error)
	Update(id int64, req model.UpdateTodoRequest) (*model.Todo, error)
	Delete(id int64) error
}

type sqliteTodoRepository struct {
	db *sql.DB
}

// NewTodoRepository returns a SQLite-backed TodoRepository.
func NewTodoRepository(db *sql.DB) TodoRepository {
	return &sqliteTodoRepository{db: db}
}

// Migrate creates the todos table if it does not exist.
func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			title       TEXT    NOT NULL,
			description TEXT    NOT NULL DEFAULT '',
			done        INTEGER NOT NULL DEFAULT 0,
			created_at  DATETIME NOT NULL,
			updated_at  DATETIME NOT NULL
		)
	`)
	return err
}

func (r *sqliteTodoRepository) FindAll() ([]model.Todo, error) {
	rows, err := r.db.Query(
		`SELECT id, title, description, done, created_at, updated_at FROM todos ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		var t model.Todo
		var done int
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &done, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.Done = done == 1
		todos = append(todos, t)
	}
	if todos == nil {
		todos = []model.Todo{}
	}
	return todos, rows.Err()
}

func (r *sqliteTodoRepository) FindByID(id int64) (*model.Todo, error) {
	row := r.db.QueryRow(
		`SELECT id, title, description, done, created_at, updated_at FROM todos WHERE id = ?`, id,
	)
	var t model.Todo
	var done int
	err := row.Scan(&t.ID, &t.Title, &t.Description, &done, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t.Done = done == 1
	return &t, nil
}

func (r *sqliteTodoRepository) Create(req model.CreateTodoRequest) (*model.Todo, error) {
	now := time.Now().UTC()
	res, err := r.db.Exec(
		`INSERT INTO todos (title, description, done, created_at, updated_at) VALUES (?, ?, 0, ?, ?)`,
		req.Title, req.Description, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *sqliteTodoRepository) Update(id int64, req model.UpdateTodoRequest) (*model.Todo, error) {
	done := 0
	if req.Done {
		done = 1
	}
	now := time.Now().UTC()
	result, err := r.db.Exec(
		`UPDATE todos SET title = ?, description = ?, done = ?, updated_at = ? WHERE id = ?`,
		req.Title, req.Description, done, now, id,
	)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, ErrNotFound
	}
	return r.FindByID(id)
}

func (r *sqliteTodoRepository) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM todos WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
