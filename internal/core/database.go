package core

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/rockpanel/rockpanel/pkg/types"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dataDir string) error {
	dbPath := filepath.Join(dataDir, "rockpanel.db")
	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)
	DB.SetConnMaxLifetime(0)
	return runMigrations()
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func runMigrations() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			command TEXT NOT NULL DEFAULT '',
			work_dir TEXT NOT NULL DEFAULT '',
			env_vars TEXT NOT NULL DEFAULT '',
			user TEXT NOT NULL DEFAULT '',
			cpu_limit INTEGER NOT NULL DEFAULT 0,
			memory_limit INTEGER NOT NULL DEFAULT 0,
			restart_policy TEXT NOT NULL DEFAULT 'on-failure',
			ports TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'stopped',
			pid INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS applications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			git_repo TEXT NOT NULL DEFAULT '',
			branch TEXT NOT NULL DEFAULT 'main',
			work_dir TEXT NOT NULL DEFAULT '',
			install_cmd TEXT NOT NULL DEFAULT '',
			build_cmd TEXT NOT NULL DEFAULT '',
			start_cmd TEXT NOT NULL DEFAULT '',
			env_vars TEXT NOT NULL DEFAULT '',
			port INTEGER NOT NULL DEFAULT 0,
			restart_policy TEXT NOT NULL DEFAULT 'on-failure',
			status TEXT NOT NULL DEFAULT 'stopped',
			pid INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS minecraft_servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			version TEXT NOT NULL DEFAULT '1.20.4',
			server_type TEXT NOT NULL DEFAULT 'vanilla',
			java_version TEXT NOT NULL DEFAULT 'java17',
			memory INTEGER NOT NULL DEFAULT 1024,
			port INTEGER NOT NULL DEFAULT 25565,
			work_dir TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'stopped',
			pid INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS backups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			target_id INTEGER NOT NULL,
			path TEXT NOT NULL DEFAULT '',
			size INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at INTEGER NOT NULL,
			completed_at INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			type TEXT NOT NULL,
			target_id INTEGER NOT NULL,
			cron_expr TEXT NOT NULL,
			command TEXT NOT NULL DEFAULT '',
			enabled INTEGER NOT NULL DEFAULT 1,
			last_run INTEGER NOT NULL DEFAULT 0,
			next_run INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS websites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT UNIQUE NOT NULL,
			target TEXT NOT NULL,
			ssl_enabled INTEGER NOT NULL DEFAULT 0,
			ssl_cert TEXT NOT NULL DEFAULT '',
			ssl_key TEXT NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS databases (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			type TEXT NOT NULL DEFAULT 'sqlite',
			host TEXT NOT NULL DEFAULT 'localhost',
			port INTEGER NOT NULL DEFAULT 0,
			username TEXT NOT NULL DEFAULT '',
			password TEXT NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS api_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			prefix TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL DEFAULT 0,
			last_used INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_servers_status ON servers(status)`,
		`CREATE INDEX IF NOT EXISTS idx_apps_status ON applications(status)`,
		`CREATE INDEX IF NOT EXISTS idx_mc_status ON minecraft_servers(status)`,
		`CREATE INDEX IF NOT EXISTS idx_backups_target ON backups(type, target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sched_enabled ON schedules(enabled)`,
		`CREATE INDEX IF NOT EXISTS idx_tokens_user ON api_tokens(user_id)`,
	}
	for _, s := range stmts {
		if _, err := DB.Exec(s); err != nil {
			return fmt.Errorf("migration: %w", err)
		}
	}
	return nil
}

func CreateDefaultAdmin(username, hashedPassword string) error {
	now := time.Now().Unix()
	_, err := DB.Exec(
		`INSERT OR IGNORE INTO users (username, password, role, created_at, updated_at) VALUES (?, ?, 'admin', ?, ?)`,
		username, hashedPassword, now, now,
	)
	return err
}

func GetUserByUsername(username string) (*types.User, error) {
	var u types.User
	err := DB.QueryRow(
		`SELECT id, username, password, role, created_at, updated_at FROM users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByID(id int64) (*types.User, error) {
	var u types.User
	err := DB.QueryRow(
		`SELECT id, username, password, role, created_at, updated_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func ListUsers() ([]types.User, error) {
	rows, err := DB.Query(`SELECT id, username, role, created_at, updated_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []types.User
	for rows.Next() {
		var u types.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func CreateUser(username, hashedPassword, role string) error {
	now := time.Now().Unix()
	_, err := DB.Exec(
		`INSERT INTO users (username, password, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		username, hashedPassword, role, now, now,
	)
	return err
}

func UpdateUser(id int64, username, role string) error {
	now := time.Now().Unix()
	_, err := DB.Exec(`UPDATE users SET username=?, role=?, updated_at=? WHERE id=?`, username, role, now, id)
	return err
}

func DeleteUser(id int64) error {
	_, err := DB.Exec(`DELETE FROM users WHERE id=?`, id)
	return err
}

func CreateAPIToken(name, tokenHash, prefix string, userID, expiresAt int64) error {
	now := time.Now().Unix()
	_, err := DB.Exec(
		`INSERT INTO api_tokens (name, token_hash, prefix, user_id, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?)`,
		name, tokenHash, prefix, userID, now, expiresAt,
	)
	return err
}

func ListAPITokens(userID int64) ([]types.APIToken, error) {
	rows, err := DB.Query(
		`SELECT id, name, prefix, user_id, created_at, expires_at, last_used FROM api_tokens WHERE user_id = ?`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tokens []types.APIToken
	for rows.Next() {
		var t types.APIToken
		if err := rows.Scan(&t.ID, &t.Name, &t.Prefix, &t.UserID, &t.CreatedAt, &t.ExpiresAt, &t.LastUsed); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

func GetAPITokenByPrefix(prefix string) (*types.APIToken, error) {
	var t types.APIToken
	err := DB.QueryRow(
		`SELECT id, name, token_hash, prefix, user_id, created_at, expires_at, last_used FROM api_tokens WHERE prefix = ?`, prefix,
	).Scan(&t.ID, &t.Name, &t.TokenHash, &t.Prefix, &t.UserID, &t.CreatedAt, &t.ExpiresAt, &t.LastUsed)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func RevokeAPIToken(id int64) error {
	_, err := DB.Exec(`DELETE FROM api_tokens WHERE id = ?`, id)
	return err
}

func UpdateAPITokenLastUsed(id int64) error {
	_, err := DB.Exec(`UPDATE api_tokens SET last_used = ? WHERE id = ?`, time.Now().Unix(), id)
	return err
}