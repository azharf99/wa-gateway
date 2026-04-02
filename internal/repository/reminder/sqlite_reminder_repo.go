package reminder

import (
	"context"
	"database/sql"

	"github.com/azharf99/wa-gateway/internal/domain"
)

type sqliteReminderRepo struct {
	db *sql.DB
}

func NewSqliteReminderRepository(db *sql.DB) domain.ReminderRepository {
	query := `
	CREATE TABLE IF NOT EXISTS reminders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target TEXT NOT NULL,
		message TEXT NOT NULL,
		is_group BOOLEAN,
		interval_days INTEGER,
		next_run DATETIME,
		is_active BOOLEAN DEFAULT 1
	);`
	db.Exec(query)
	return &sqliteReminderRepo{db: db}
}

// Implementasi Create, GetAll, Update, Delete standard...
func (r *sqliteReminderRepo) Create(ctx context.Context, rem *domain.Reminder) error {
	query := `INSERT INTO reminders (target, message, is_group, interval_days, next_run) VALUES (?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, rem.To, rem.Message, rem.IsGroup, rem.IntervalDays, rem.NextRun)
	if err != nil {
		return err
	}
	rem.ID, _ = res.LastInsertId()
	return nil
}

func (r *sqliteReminderRepo) GetAll(ctx context.Context) ([]domain.Reminder, error) {
	rows, _ := r.db.QueryContext(ctx, "SELECT id, target, message, is_group, interval_days, next_run, is_active FROM reminders")
	defer rows.Close()
	var list []domain.Reminder
	for rows.Next() {
		var rem domain.Reminder
		rows.Scan(&rem.ID, &rem.To, &rem.Message, &rem.IsGroup, &rem.IntervalDays, &rem.NextRun, &rem.IsActive)
		list = append(list, rem)
	}
	return list, nil
}

func (r *sqliteReminderRepo) Update(ctx context.Context, rem *domain.Reminder) error {
	query := `UPDATE reminders SET target=?, message=?, is_group=?, interval_days=?, next_run=?, is_active=? WHERE id=?`
	_, err := r.db.ExecContext(ctx, query, rem.To, rem.Message, rem.IsGroup, rem.IntervalDays, rem.NextRun, rem.IsActive, rem.ID)
	return err
}

func (r *sqliteReminderRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM reminders WHERE id=?", id)
	return err
}

func (r *sqliteReminderRepo) GetByID(ctx context.Context, id int64) (*domain.Reminder, error) {
	var rem domain.Reminder
	err := r.db.QueryRowContext(ctx, "SELECT id, target, message, is_group, interval_days, next_run, is_active FROM reminders WHERE id=?", id).
		Scan(&rem.ID, &rem.To, &rem.Message, &rem.IsGroup, &rem.IntervalDays, &rem.NextRun, &rem.IsActive)
	return &rem, err
}
