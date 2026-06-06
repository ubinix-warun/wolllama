// Package db manages the SQLite database for the Wolllama API.
package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps a sql.DB connection with wolllama-specific queries.
type DB struct {
	conn *sql.DB
}

// Model represents a submitted model in the registry.
type Model struct {
	ID            int64     `json:"id"`
	SubmitterID   int64     `json:"submitter_id"`
	SubmitterName    string    `json:"submitter_name,omitempty"`
	AvatarURL        *string   `json:"avatar_url,omitempty"`
	WalletAddress    *string   `json:"wallet_address,omitempty"`
	ManifestObjID    string    `json:"manifest_obj_id"`
	DisplayName   string    `json:"display_name"`
	DescriptionMd *string   `json:"description_md,omitempty"`
	OriginalName  *string   `json:"original_name,omitempty"`
	Tag           *string   `json:"tag,omitempty"`
	TotalSize     *int64    `json:"total_size,omitempty"`
	BlobCount     *int      `json:"blob_count,omitempty"`
	ManifestJSON     *string   `json:"manifest_json,omitempty"`
	SubmitterAddress *string   `json:"submitter_address,omitempty"`
	Signature        *string   `json:"signature,omitempty"`
	Available        bool      `json:"available"`
	CreatedAt     time.Time `json:"created_at"`
}

// User represents a registered user.
type User struct {
	ID            int64     `json:"id"`
	GitHubID      int64     `json:"github_id"`
	Username      string    `json:"username"`
	AvatarURL     *string   `json:"avatar_url,omitempty"`
	WalletAddress *string   `json:"wallet_address,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// Open opens a SQLite database at the given path.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	conn.SetMaxOpenConns(1) // SQLite serializes writes

	return &DB{conn: conn}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Migrate creates the initial schema if it doesn't exist.
func (db *DB) Migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		github_id INTEGER DEFAULT 0,
		username TEXT NOT NULL,
		avatar_url TEXT,
		wallet_address TEXT UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS models (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		submitter_id INTEGER NOT NULL REFERENCES users(id),
		manifest_obj_id TEXT UNIQUE NOT NULL,
		display_name TEXT NOT NULL,
		description_md TEXT,
		original_name TEXT,
		tag TEXT,
		total_size INTEGER,
		blob_count INTEGER,
		manifest_json TEXT,
		submitter_address TEXT,
		signature TEXT,
		available BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_models_display_name ON models(display_name);
	CREATE INDEX IF NOT EXISTS idx_models_submitter ON models(submitter_id);
	CREATE INDEX IF NOT EXISTS idx_models_available ON models(available);
	`
	_, err := db.conn.Exec(schema)
	return err
}

// GetOrCreateAnonUser returns the anonymous user, creating one if it doesn't exist.
func (db *DB) GetOrCreateAnonUser() (*User, error) {
	// Try to find existing anonymous user
	user := &User{}
	err := db.conn.QueryRow(`
		SELECT id, github_id, username, avatar_url, wallet_address, created_at
		FROM users WHERE github_id = 0 AND username = 'anonymous'
		LIMIT 1
	`).Scan(&user.ID, &user.GitHubID, &user.Username, &user.AvatarURL, &user.WalletAddress, &user.CreatedAt)

	if err == nil {
		return user, nil
	}

	// Create one
	err = db.conn.QueryRow(`
		INSERT INTO users (github_id, username)
		VALUES (0, 'anonymous')
		RETURNING id, github_id, username, avatar_url, wallet_address, created_at
	`).Scan(&user.ID, &user.GitHubID, &user.Username, &user.AvatarURL, &user.WalletAddress, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create anon user: %w", err)
	}
	return user, nil
}

// CreateUserByWallet creates or looks up a user by Sui wallet address.
func (db *DB) CreateUserByWallet(address string) (*User, error) {
	username := "sui:" + address[:12] + "..."
	user := &User{}
	err := db.conn.QueryRow(`
		INSERT INTO users (github_id, username, wallet_address)
		VALUES (0, ?, ?)
		ON CONFLICT(wallet_address) DO UPDATE SET username=excluded.username
		RETURNING id, github_id, username, avatar_url, wallet_address, created_at
	`, username, address).Scan(&user.ID, &user.GitHubID, &user.Username, &user.AvatarURL, &user.WalletAddress, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user by wallet: %w", err)
	}
	return user, nil
}

// CreateUser inserts a new user or returns the existing one.
func (db *DB) CreateUser(githubID int64, username string, avatarURL *string) (*User, error) {
	user := &User{}
	err := db.conn.QueryRow(`
		INSERT INTO users (github_id, username, avatar_url)
		VALUES (?, ?, ?)
		ON CONFLICT(github_id) DO UPDATE SET username=excluded.username, avatar_url=excluded.avatar_url
		RETURNING id, github_id, username, avatar_url, created_at
	`, githubID, username, avatarURL).Scan(&user.ID, &user.GitHubID, &user.Username, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// GetUserByGitHubID looks up a user by their GitHub ID.
func (db *DB) GetUserByGitHubID(githubID int64) (*User, error) {
	user := &User{}
	err := db.conn.QueryRow(`
		SELECT id, github_id, username, avatar_url, created_at
		FROM users WHERE github_id = ?
	`, githubID).Scan(&user.ID, &user.GitHubID, &user.Username, &user.AvatarURL, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

// GetUserByID looks up a user by their database ID.
func (db *DB) GetUserByID(id int64) (*User, error) {
	user := &User{}
	err := db.conn.QueryRow(`
		SELECT id, github_id, username, avatar_url, created_at
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.GitHubID, &user.Username, &user.AvatarURL, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

// CreateModel inserts a new model submission.
func (db *DB) CreateModel(m *Model) error {
	err := db.conn.QueryRow(`
		INSERT INTO models (submitter_id, manifest_obj_id, display_name, description_md, original_name, tag, total_size, blob_count, manifest_json, submitter_address, signature)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at
	`, m.SubmitterID, m.ManifestObjID, m.DisplayName, m.DescriptionMd, m.OriginalName, m.Tag, m.TotalSize, m.BlobCount, m.ManifestJSON, m.SubmitterAddress, m.Signature).
		Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return fmt.Errorf("create model: %w", err)
	}
	return nil
}

// GetModelByID fetches a single model by its database ID.
func (db *DB) GetModelByID(id int64) (*Model, error) {
	m := &Model{}
	err := db.conn.QueryRow(`
		SELECT m.id, m.submitter_id, u.username, u.avatar_url, u.wallet_address,
			   m.manifest_obj_id, m.display_name, m.description_md,
			   m.original_name, m.tag, m.total_size, m.blob_count,
			   m.manifest_json, m.submitter_address, m.signature, m.available, m.created_at
		FROM models m
		JOIN users u ON u.id = m.submitter_id
		WHERE m.id = ?
	`, id).Scan(
		&m.ID, &m.SubmitterID, &m.SubmitterName, &m.AvatarURL, &m.WalletAddress,
		&m.ManifestObjID, &m.DisplayName, &m.DescriptionMd,
		&m.OriginalName, &m.Tag, &m.TotalSize, &m.BlobCount,
		&m.ManifestJSON, &m.SubmitterAddress, &m.Signature, &m.Available, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get model: %w", err)
	}
	return m, nil
}

// ListModels returns a paginated list of available models.
func (db *DB) ListModels(offset, limit int, search string) ([]Model, error) {
	query := `
		SELECT m.id, m.submitter_id, u.username, u.avatar_url, u.wallet_address,
			   m.manifest_obj_id, m.display_name, m.description_md,
			   m.original_name, m.tag, m.total_size, m.blob_count,
			   m.manifest_json, m.submitter_address, m.signature, m.available, m.created_at
		FROM models m
		JOIN users u ON u.id = m.submitter_id
		WHERE m.available = 1
	`
	args := []interface{}{}
	if search != "" {
		query += ` AND m.display_name LIKE ?`
		args = append(args, "%"+search+"%")
	}
	query += ` ORDER BY m.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var m Model
		if err := rows.Scan(
			&m.ID, &m.SubmitterID, &m.SubmitterName, &m.AvatarURL, &m.WalletAddress,
			&m.ManifestObjID, &m.DisplayName, &m.DescriptionMd,
			&m.OriginalName, &m.Tag, &m.TotalSize, &m.BlobCount,
			&m.ManifestJSON, &m.SubmitterAddress, &m.Signature, &m.Available, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan model: %w", err)
		}
		models = append(models, m)
	}
	return models, rows.Err()
}

// ListModelsByUser returns models submitted by a specific user.
func (db *DB) ListModelsByUser(userID int64) ([]Model, error) {
	rows, err := db.conn.Query(`
		SELECT m.id, m.submitter_id, u.username, u.avatar_url, u.wallet_address,
			   m.manifest_obj_id, m.display_name, m.description_md,
			   m.original_name, m.tag, m.total_size, m.blob_count,
			   m.manifest_json, m.submitter_address, m.signature, m.available, m.created_at
		FROM models m
		JOIN users u ON u.id = m.submitter_id
		WHERE m.submitter_id = ?
		ORDER BY m.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user models: %w", err)
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var m Model
		if err := rows.Scan(
			&m.ID, &m.SubmitterID, &m.SubmitterName, &m.AvatarURL, &m.WalletAddress,
			&m.ManifestObjID, &m.DisplayName, &m.DescriptionMd,
			&m.OriginalName, &m.Tag, &m.TotalSize, &m.BlobCount,
			&m.ManifestJSON, &m.SubmitterAddress, &m.Signature, &m.Available, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan model: %w", err)
		}
		models = append(models, m)
	}
	return models, rows.Err()
}

// SetModelUnavailable marks a model as unavailable (blobs expired on Walrus).
func (db *DB) SetModelUnavailable(id int64) error {
	_, err := db.conn.Exec(`UPDATE models SET available = 0 WHERE id = ?`, id)
	return err
}

// ListAllModelObjIDs returns all manifest object IDs for health checking.
func (db *DB) ListAllModelObjIDs() ([]struct {
	ID            int64
	ManifestObjID string
}, error) {
	rows, err := db.conn.Query(`SELECT id, manifest_obj_id FROM models WHERE available = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		ID            int64
		ManifestObjID string
	}
	for rows.Next() {
		var r struct {
			ID            int64
			ManifestObjID string
		}
		if err := rows.Scan(&r.ID, &r.ManifestObjID); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
