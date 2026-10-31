// Package store 负责把微信读书凭据持久化到 SQLite。
// 账号以 vid(微信读书用户 ID)为唯一键。
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// ErrNotFound 表示该账号不存在。
var ErrNotFound = errors.New("账号不存在")

const schema = `
CREATE TABLE IF NOT EXISTS weread_account (
  vid           TEXT PRIMARY KEY,
  refresh_token TEXT NOT NULL,
  device_id     TEXT NOT NULL,
  access_token  TEXT NOT NULL DEFAULT '',
  remark        TEXT NOT NULL DEFAULT '',
  rotated_at    INTEGER NOT NULL,
  created_at    INTEGER NOT NULL
);`

// DefaultPath 返回默认数据库路径:当前目录下的 data/weread.db,
// 即在项目根目录运行时落在 <项目根>/data/ 下。相对路径,随工作目录走。
func DefaultPath() (string, error) {
	return filepath.Join("data", "weread.db"), nil
}

// Open 打开(必要时创建)数据库并执行迁移。
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}
	// modernc.org/sqlite 的 DSN 通过 _pragma 传连接级编译指令。
	// 相对路径必须先转绝对:否则 url.URL 会把首段当成 URI authority(file://data/...),
	// 驱动直接报 invalid uri authority。
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("解析数据库路径失败: %w", err)
	}
	dsn := (&url.URL{Scheme: "file", Path: abs}).String() + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("启用外键约束失败: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}
	// 这个文件等于账号控制权,SQLite 新建文件不会自动收紧权限,WAL 伴生文件同理。
	for _, p := range []string{path, path + "-wal", path + "-shm"} {
		_ = os.Chmod(p, 0o600)
	}
	return db, nil
}

// migrate 在同一事务内升级结构和版本，失败后保留旧版本。
func migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS weread_schema (version INTEGER PRIMARY KEY)`); err != nil {
		return err
	}
	var version int
	if err := tx.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM weread_schema`).Scan(&version); err != nil {
		return err
	}
	if version > 1 {
		return fmt.Errorf("数据库版本 %d 高于程序支持的版本 1", version)
	}
	// 基线迁移必须对已有库也执行:新增列通过 ensureColumn 幂等补齐,
	// 否则升级到带新列的程序后旧库会报 "no such column"。
	if err := migrateBaseline(tx); err != nil {
		return err
	}
	if version == 0 {
		if _, err := tx.Exec(`INSERT INTO weread_schema(version) VALUES (1)`); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type schemaDB interface {
	Exec(string, ...any) (sql.Result, error)
	Query(string, ...any) (*sql.Rows, error)
}

func migrateBaseline(db schemaDB) error {
	for _, stmt := range []string{schema, readingSchema, readingIndexes, logsSchema, pushSchema, settingsSchema} {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	// 补齐当前版本字段，便于从空数据库逐步创建完整结构。
	for _, column := range []struct{ table, name, def string }{
		{"weread_account", "remark", "TEXT NOT NULL DEFAULT ''"},
		{"weread_account", "name", "TEXT NOT NULL DEFAULT ''"},
		{"weread_account", "avatar", "TEXT NOT NULL DEFAULT ''"},
		{"weread_account", "user_vid", "TEXT NOT NULL DEFAULT ''"},
		{"weread_account", "profile", "TEXT NOT NULL DEFAULT ''"},
		{"weread_account", "card", "TEXT NOT NULL DEFAULT ''"},
		{"weread_account", "shelf", "TEXT NOT NULL DEFAULT ''"},
		{"weread_account", "details_cached_at", "INTEGER NOT NULL DEFAULT 0"},
		{"weread_account", "profile_updated_at", "INTEGER NOT NULL DEFAULT 0"},
	} {
		if err := ensureColumn(db, column.table, column.name, column.def); err != nil {
			return err
		}
	}
	return nil
}

func ensureColumn(db schemaDB, table, column, definition string) error {
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var cid, notNull, pk int
		var name, ctype string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		if name == column {
			found = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if found {
		return nil
	}
	_, err = db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + definition)
	return err
}

// Credential 是 weread_account 表一行的内存表示。
type Credential struct {
	Vid          string
	RefreshToken string
	DeviceID     string
	AccessToken  string
	Remark       string
	Name         string
	Avatar       string
	UserVid      string
	// 详情缓存:三段原始 JSON 与采集时间;空串表示还没有缓存。
	Profile          string
	Card             string
	Shelf            string
	DetailsCachedAt  int64
	ProfileUpdatedAt int64
	RotatedAt        time.Time
	CreatedAt        time.Time
}

const selectCols = `vid, refresh_token, device_id, access_token, remark, name, avatar, user_vid, profile, card, shelf, rotated_at, created_at, details_cached_at, profile_updated_at`

func scanCredential(scan func(...any) error) (*Credential, error) {
	var c Credential
	var rotated, created int64
	if err := scan(&c.Vid, &c.RefreshToken, &c.DeviceID, &c.AccessToken, &c.Remark, &c.Name, &c.Avatar, &c.UserVid, &c.Profile, &c.Card, &c.Shelf, &rotated, &created, &c.DetailsCachedAt, &c.ProfileUpdatedAt); err != nil {
		return nil, err
	}
	c.RotatedAt = time.Unix(rotated, 0)
	c.CreatedAt = time.Unix(created, 0)
	return &c, nil
}

func Load(db *sql.DB, vid string) (*Credential, error) {
	row := db.QueryRow(`SELECT `+selectCols+` FROM weread_account WHERE vid = ?`, vid)
	c, err := scanCredential(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

// List 返回全部账号,按创建时间排序。
func List(db *sql.DB) ([]*Credential, error) {
	rows, err := db.Query(`SELECT ` + selectCols + ` FROM weread_account ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Credential
	for rows.Next() {
		c, err := scanCredential(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Save 按 vid UPSERT 凭据,rotated_at 记录本次写入时间。
// 每次续期拿到新 refreshToken 后都必须立刻调用,否则丢了轮换值会话就断了。
// 注意冲突分支不更新 remark/name:token 轮换不能清掉用户设置的展示信息。
func Save(db *sql.DB, c *Credential) error {
	if c.Vid == "" || c.RefreshToken == "" || c.DeviceID == "" {
		return errors.New("凭据字段不完整,拒绝写入")
	}
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO weread_account (vid, refresh_token, device_id, access_token, remark, rotated_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(vid) DO UPDATE SET
			refresh_token = excluded.refresh_token,
			device_id = excluded.device_id,
			access_token = excluded.access_token,
			rotated_at = excluded.rotated_at`,
		c.Vid, c.RefreshToken, c.DeviceID, c.AccessToken, c.Remark, now, now)
	return err
}

// UpdateRemark 更新账号备注(界面上的显示名)。
func UpdateRemark(db *sql.DB, vid, remark string) error {
	res, err := db.Exec(`UPDATE weread_account SET remark = ? WHERE vid = ?`, remark, vid)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete 删除账号记录(不影响微信读书服务端的会话)。
func Delete(db *sql.DB, vid string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`DELETE FROM weread_account WHERE vid = ?`, vid)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	if _, err := tx.Exec(`DELETE FROM weread_reading WHERE vid = ?`, vid); err != nil {
		return err
	}
	// 历史日志一并清理,避免日志页残留已删除账号的条目。
	if _, err := tx.Exec(`DELETE FROM weread_log WHERE vid = ?`, vid); err != nil {
		return err
	}
	return tx.Commit()
}
