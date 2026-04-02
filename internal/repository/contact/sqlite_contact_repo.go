package contact

import (
	"context"
	"database/sql"

	"github.com/azharf99/wa-gateway/internal/domain"
)

type sqliteContactRepo struct {
	db *sql.DB
}

func NewSqliteContactRepository(db *sql.DB) domain.ContactRepository {
	query := `
	CREATE TABLE IF NOT EXISTS contacts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		phone TEXT UNIQUE NOT NULL,
		category TEXT
	);`
	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
	return &sqliteContactRepo{db: db}
}

func (r *sqliteContactRepo) Create(ctx context.Context, c *domain.Contact) error {
	query := `INSERT INTO contacts (name, phone, category) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, c.Name, c.Phone, c.Category)
	return err
}

func (r *sqliteContactRepo) GetAll(ctx context.Context) ([]domain.Contact, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, phone, category FROM contacts ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []domain.Contact
	for rows.Next() {
		var c domain.Contact
		rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Category)
		contacts = append(contacts, c)
	}
	return contacts, nil
}

func (r *sqliteContactRepo) GetByPhone(ctx context.Context, phone string) (*domain.Contact, error) {
	var c domain.Contact
	err := r.db.QueryRowContext(ctx, "SELECT id, name, phone, category FROM contacts WHERE phone = ?", phone).
		Scan(&c.ID, &c.Name, &c.Phone, &c.Category)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *sqliteContactRepo) Update(ctx context.Context, c *domain.Contact) error {
	_, err := r.db.ExecContext(ctx, "UPDATE contacts SET name = ?, phone = ?, category = ? WHERE id = ?", c.Name, c.Phone, c.Category, c.ID)
	return err
}

func (r *sqliteContactRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM contacts WHERE id = ?", id)
	return err
}

func (r *sqliteContactRepo) ImportCSV(ctx context.Context, contacts []domain.Contact) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Gunakan format "INSERT OR REPLACE" agar jika ada nomor yang sama, data terbaru yang diambil
	stmt, err := tx.PrepareContext(ctx, "INSERT OR REPLACE INTO contacts (name, phone, category) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, c := range contacts {
		_, err := stmt.ExecContext(ctx, c.Name, c.Phone, c.Category)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
