package store

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type SQLiteStore struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Enable WAL mode for concurrent read/write
	dsn := fmt.Sprintf("file:%s?mode=rwc&cache=shared&_journal_mode=WAL&_busy_timeout=5000", dbPath)
	
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	// Create tables if they don't exist
	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// Verify WAL mode
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		log.Warn().Err(err).Msg("could not verify journal mode")
	} else {
		log.Info().Str("journal_mode", journalMode).Msg("database initialized")
	}

	return &SQLiteStore{db: db}, nil
}

func createSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS credentials (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		type TEXT NOT NULL,
		token TEXT NOT NULL,
		metadata TEXT,
		created_at DATETIME NOT NULL,
		last_used DATETIME,
		usage_count INTEGER DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		action TEXT NOT NULL,
		credential_name TEXT,
		source_ip TEXT,
		source_tool TEXT,
		status TEXT,
		details TEXT
	);
	`
	_, err := db.Exec(query)
	return err
}

func (s *SQLiteStore) GetCredential(ctx context.Context, name string) (*Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cred Credential
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, type, token, metadata, created_at, last_used, usage_count
		 FROM credentials WHERE name = ?`,
		name).Scan(&cred.ID, &cred.Name, &cred.Type, &cred.Token, &cred.Metadata,
		&cred.CreatedAt, &cred.LastUsed, &cred.UsageCount)

	if err == sql.ErrNoRows {
		return nil, ErrCredentialNotFound
	}
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (s *SQLiteStore) AddCredential(ctx context.Context, cred *Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO credentials (id, name, type, token, metadata, created_at, last_used, usage_count)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		cred.ID, cred.Name, cred.Type, cred.Token, cred.Metadata, cred.CreatedAt, cred.LastUsed, cred.UsageCount)
	return err
}

func (s *SQLiteStore) UpdateCredential(ctx context.Context, cred *Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx,
		`UPDATE credentials SET type = ?, token = ?, metadata = ?
		 WHERE name = ?`,
		cred.Type, cred.Token, cred.Metadata, cred.Name)
	return err
}

func (s *SQLiteStore) DeleteCredential(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, "DELETE FROM credentials WHERE name = ?", name)
	return err
}

func (s *SQLiteStore) ListCredentials(ctx context.Context) ([]*Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx, "SELECT id, name, type, token, metadata, created_at, last_used, usage_count FROM credentials")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*Credential
	for rows.Next() {
		var cred Credential
		if err := rows.Scan(&cred.ID, &cred.Name, &cred.Type, &cred.Token, &cred.Metadata, &cred.CreatedAt, &cred.LastUsed, &cred.UsageCount); err != nil {
			return nil, err
		}
		results = append(results, &cred)
	}
	return results, nil
}

func (s *SQLiteStore) UpdateLastUsed(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx,
		"UPDATE credentials SET last_used = ?, usage_count = usage_count + 1 WHERE name = ?",
		time.Now(), name)
	return err
}

func (s *SQLiteStore) AddAuditLog(ctx context.Context, entry *AuditLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_logs (timestamp, action, credential_name, source_ip, source_tool, status, details)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		time.Now(), entry.Action, entry.CredentialName, entry.SourceIP, entry.SourceTool, entry.Status, entry.Details)
	return err
}

func (s *SQLiteStore) ListAuditLogs(ctx context.Context, name string, limit int) ([]*AuditLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, timestamp, action, credential_name, source_ip, source_tool, status, details FROM audit_logs`
	var args []interface{}
	if name != "" {
		query += " WHERE credential_name = ?"
		args = append(args, name)
	}
	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.Timestamp, &log.Action, &log.CredentialName, &log.SourceIP, &log.SourceTool, &log.Status, &log.Details); err != nil {
			return nil, err
		}
		results = append(results, &log)
	}
	return results, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
