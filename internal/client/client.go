package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(baseURL, apiKey string, insecureSkipVerify bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify}, //nolint:gosec
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Transport: transport,
		},
	}
}

type apiError struct {
	Detail interface{} `json:"detail"`
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr apiError
		if jsonErr := json.Unmarshal(respBody, &apiErr); jsonErr == nil && apiErr.Detail != nil {
			return nil, resp.StatusCode, fmt.Errorf("API error %d: %v", resp.StatusCode, apiErr.Detail)
		}
		return nil, resp.StatusCode, fmt.Errorf("API error %d: %s", resp.StatusCode, resp.Status)
	}

	return respBody, resp.StatusCode, nil
}

// ServerConfig represents a barman server configuration.
type ServerConfig struct {
	Name                       string  `json:"name,omitempty"`
	Description                *string `json:"description"`
	Conninfo                   string  `json:"conninfo"`
	SSHCommand                 *string `json:"ssh_command"`
	BackupMethod               string  `json:"backup_method"`
	Archiver                   bool    `json:"archiver"`
	StreamingConninfo          *string `json:"streaming_conninfo"`
	StreamingArchiver          bool    `json:"streaming_archiver"`
	CreateSlot                 string  `json:"create_slot"`
	SlotName                   *string `json:"slot_name"`
	PathPrefix                 *string `json:"path_prefix"`
	SSLMode                    string  `json:"sslmode"`
	RetentionPolicy            *string `json:"retention_policy"`
	WALRetentionPolicy         string  `json:"wal_retention_policy"`
	MinimumRedundancy          int64   `json:"minimum_redundancy"`
	Compression                *string `json:"compression"`
	BackupCompression          *string `json:"backup_compression"`
	StreamingArchiverBatchSize int64   `json:"streaming_archiver_batch_size"`
	PGVersion                  int64   `json:"pg_version"`
	BackupsEnabled             bool    `json:"backups_enabled"`
}

type ServerConfigCreateRequest struct {
	ServerConfig
	ScheduleEnabled *bool `json:"schedule_enabled,omitempty"`
}

func (c *Client) GetServerConfig(ctx context.Context, name string) (*ServerConfig, error) {
	data, status, err := c.do(ctx, http.MethodGet, "/api/sql/server-configs/"+name, nil)
	if err != nil {
		if status == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	var sc ServerConfig
	if err := json.Unmarshal(data, &sc); err != nil {
		return nil, fmt.Errorf("decode server config: %w", err)
	}
	return &sc, nil
}

func (c *Client) CreateServerConfig(ctx context.Context, req ServerConfigCreateRequest) (*ServerConfig, error) {
	data, _, err := c.do(ctx, http.MethodPost, "/api/sql/server-configs", req)
	if err != nil {
		return nil, err
	}
	var sc ServerConfig
	if err := json.Unmarshal(data, &sc); err != nil {
		return nil, fmt.Errorf("decode server config: %w", err)
	}
	return &sc, nil
}

func (c *Client) UpdateServerConfig(ctx context.Context, name string, req ServerConfig) (*ServerConfig, error) {
	data, _, err := c.do(ctx, http.MethodPut, "/api/sql/server-configs/"+name, req)
	if err != nil {
		return nil, err
	}
	var sc ServerConfig
	if err := json.Unmarshal(data, &sc); err != nil {
		return nil, fmt.Errorf("decode server config: %w", err)
	}
	return &sc, nil
}

func (c *Client) SetBackupsEnabled(ctx context.Context, name string, enabled bool) error {
	body := map[string]bool{"enabled": enabled}
	_, _, err := c.do(ctx, http.MethodPatch, "/api/sql/server-configs/"+name+"/backups-enabled", body)
	return err
}

func (c *Client) DeleteServerConfig(ctx context.Context, name string) error {
	_, status, err := c.do(ctx, http.MethodDelete, "/api/sql/server-configs/"+name, nil)
	if err != nil {
		if status == http.StatusNotFound {
			return nil
		}
		return err
	}
	return nil
}

// Schedule represents a backup schedule.
type Schedule struct {
	ID             int64          `json:"id"`
	ServerName     string         `json:"server_name"`
	Label          *string        `json:"label"`
	ScheduleType   string         `json:"schedule_type"`
	ScheduleConfig ScheduleConfig `json:"schedule_config"`
	Enabled        bool           `json:"enabled"`
	NextRunAt      *string        `json:"next_run_at"`
	LastRunAt      *string        `json:"last_run_at"`
	CreatedAt      string         `json:"created_at"`
}

type ScheduleConfig struct {
	RunAt *string `json:"run_at,omitempty"`
	Time  *string `json:"time,omitempty"`
	Days  []int64 `json:"days,omitempty"`
	Day   *int64  `json:"day,omitempty"`
}

type ScheduleRequest struct {
	ServerName     string         `json:"server_name"`
	Label          *string        `json:"label"`
	ScheduleType   string         `json:"schedule_type"`
	ScheduleConfig ScheduleConfig `json:"schedule_config"`
	Enabled        bool           `json:"enabled"`
}

func (c *Client) GetSchedule(ctx context.Context, id int64) (*Schedule, error) {
	data, status, err := c.do(ctx, http.MethodGet, "/api/sql/schedules/"+strconv.FormatInt(id, 10), nil)
	if err != nil {
		if status == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	var s Schedule
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("decode schedule: %w", err)
	}
	return &s, nil
}

func (c *Client) CreateSchedule(ctx context.Context, req ScheduleRequest) (*Schedule, error) {
	data, _, err := c.do(ctx, http.MethodPost, "/api/sql/schedules", req)
	if err != nil {
		return nil, err
	}
	var s Schedule
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("decode schedule: %w", err)
	}
	return &s, nil
}

func (c *Client) UpdateSchedule(ctx context.Context, id int64, req ScheduleRequest) (*Schedule, error) {
	data, _, err := c.do(ctx, http.MethodPut, "/api/sql/schedules/"+strconv.FormatInt(id, 10), req)
	if err != nil {
		return nil, err
	}
	var s Schedule
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("decode schedule: %w", err)
	}
	return &s, nil
}

func (c *Client) DeleteSchedule(ctx context.Context, id int64) error {
	_, status, err := c.do(ctx, http.MethodDelete, "/api/sql/schedules/"+strconv.FormatInt(id, 10), nil)
	if err != nil {
		if status == http.StatusNotFound {
			return nil
		}
		return err
	}
	return nil
}

// APIKey represents an API key.
type APIKey struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	KeyPrefix  string  `json:"key_prefix"`
	Role       string  `json:"role"`
	CreatedBy  string  `json:"created_by"`
	LastUsedAt *string `json:"last_used_at"`
	CreatedAt  string  `json:"created_at"`
	Key        *string `json:"key,omitempty"`
}

type APIKeyCreateRequest struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

func (c *Client) ListAPIKeys(ctx context.Context) ([]APIKey, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/api/sql/api-keys", nil)
	if err != nil {
		return nil, err
	}
	var keys []APIKey
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("decode api keys: %w", err)
	}
	return keys, nil
}

func (c *Client) CreateAPIKey(ctx context.Context, req APIKeyCreateRequest) (*APIKey, error) {
	data, _, err := c.do(ctx, http.MethodPost, "/api/sql/api-keys", req)
	if err != nil {
		return nil, err
	}
	var key APIKey
	if err := json.Unmarshal(data, &key); err != nil {
		return nil, fmt.Errorf("decode api key: %w", err)
	}
	return &key, nil
}

func (c *Client) DeleteAPIKey(ctx context.Context, id int64) error {
	_, status, err := c.do(ctx, http.MethodDelete, "/api/sql/api-keys/"+strconv.FormatInt(id, 10), nil)
	if err != nil {
		if status == http.StatusNotFound {
			return nil
		}
		return err
	}
	return nil
}

// ServerStatus represents the status of a barman server.
type ServerStatus struct {
	OK             bool               `json:"ok"`
	CheckOK        bool               `json:"check_ok"`
	CheckItems     []CheckItem        `json:"check_items"`
	ReplicationJSON *string
	Fields         map[string]interface{}
}

type CheckItem struct {
	Check  string `json:"check"`
	Status string `json:"status"`
	Hint   string `json:"hint"`
}

func (c *Client) GetServerStatus(ctx context.Context, serverName string) (*ServerStatus, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/api/sql/servers/"+serverName+"/status", nil)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("decode server status: %w", err)
	}

	ss := &ServerStatus{
		Fields: make(map[string]interface{}),
	}

	if v, ok := raw["ok"].(bool); ok {
		ss.OK = v
	}
	if v, ok := raw["check_ok"].(bool); ok {
		ss.CheckOK = v
	}

	if items, ok := raw["check_items"].([]interface{}); ok {
		for _, item := range items {
			if m, ok := item.(map[string]interface{}); ok {
				ci := CheckItem{}
				if v, ok := m["check"].(string); ok {
					ci.Check = v
				}
				if v, ok := m["status"].(string); ok {
					ci.Status = v
				}
				if v, ok := m["hint"].(string); ok {
					ci.Hint = v
				}
				ss.CheckItems = append(ss.CheckItems, ci)
			}
		}
	}

	if v, ok := raw["replication_json"]; ok && v != nil {
		if b, err := json.Marshal(v); err == nil {
			s := string(b)
			ss.ReplicationJSON = &s
		}
	}

	reserved := map[string]bool{"ok": true, "check_ok": true, "check_items": true, "replication_json": true}
	for k, v := range raw {
		if reserved[k] {
			continue
		}
		if s, ok := v.(string); ok {
			ss.Fields[k] = s
		}
	}

	return ss, nil
}

// Backup represents a barman backup entry.
type Backup struct {
	BackupID   string  `json:"backup_id"`
	Status     string  `json:"status"`
	Size       string  `json:"size"`
	BeginTime  *string `json:"begin_time"`
	EndTime    *string `json:"end_time"`
	BackupType *string `json:"backup_type"`
	Source     *string `json:"source"`
}

func (c *Client) ListBackups(ctx context.Context, serverName string) ([]Backup, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/api/sql/servers/"+serverName+"/backups", nil)
	if err != nil {
		return nil, err
	}
	var backups []Backup
	if err := json.Unmarshal(data, &backups); err != nil {
		return nil, fmt.Errorf("decode backups: %w", err)
	}
	return backups, nil
}

// Job represents a command queue entry.
type Job struct {
	ID          int64   `json:"id"`
	CacheKey    *string `json:"cache_key"`
	ArgsJSON    string  `json:"args_json"`
	Status      string  `json:"status"`
	ExitCode    *int64  `json:"exit_code"`
	Stdout      *string `json:"stdout"`
	Stderr      *string `json:"stderr"`
	QueuedAt    string  `json:"queued_at"`
	StartedAt   *string `json:"started_at"`
	CompletedAt *string `json:"completed_at"`
	ScheduleID  *int64  `json:"schedule_id"`
	PGVersion   *int64  `json:"pg_version"`
}

func (c *Client) ListJobs(ctx context.Context, server, status string, limit int64) ([]Job, error) {
	params := url.Values{}
	if server != "" {
		params.Set("server", server)
	}
	if status != "" {
		params.Set("status", status)
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	path := "/api/sql/jobs"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	data, _, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var jobs []Job
	if err := json.Unmarshal(data, &jobs); err != nil {
		return nil, fmt.Errorf("decode jobs: %w", err)
	}
	return jobs, nil
}

// SSHKey represents the SSH public key response.
type SSHKey struct {
	PublicKey *string `json:"public_key"`
}

func (c *Client) GetSSHKey(ctx context.Context) (*SSHKey, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/api/sql/ssh-key", nil)
	if err != nil {
		return nil, err
	}
	var key SSHKey
	if err := json.Unmarshal(data, &key); err != nil {
		return nil, fmt.Errorf("decode ssh key: %w", err)
	}
	return &key, nil
}

// ServerSummary is returned by GET /servers.
type ServerSummary struct {
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	Active          bool    `json:"active"`
	BackupCount     int64   `json:"backup_count"`
	LastBackup      *string `json:"last_backup"`
	DiskUsage       *string `json:"disk_usage"`
	DiskBytes       *int64  `json:"disk_bytes"`
	RetentionPolicy *string `json:"retention_policy"`
	CheckOK         bool    `json:"check_ok"`
	RedundancyOK    bool    `json:"redundancy_ok"`
	RedundancyRaw   *string `json:"redundancy_raw"`
	HasActiveBackup bool    `json:"has_active_backup"`
}

func (c *Client) ListServers(ctx context.Context) ([]ServerSummary, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/api/sql/servers", nil)
	if err != nil {
		return nil, err
	}
	var servers []ServerSummary
	if err := json.Unmarshal(data, &servers); err != nil {
		return nil, fmt.Errorf("decode servers: %w", err)
	}
	return servers, nil
}

func (c *Client) ListServerConfigs(ctx context.Context) ([]ServerConfig, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/api/sql/server-configs", nil)
	if err != nil {
		return nil, err
	}
	var configs []ServerConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("decode server configs: %w", err)
	}
	return configs, nil
}

// Stats is returned by GET /stats.
type Stats struct {
	TotalBackups    int64  `json:"total_backups"`
	TotalDisk       string `json:"total_disk"`
	BarmanDiskTotal *int64 `json:"barman_disk_total"`
	BarmanDiskFree  *int64 `json:"barman_disk_free"`
}

func (c *Client) GetStats(ctx context.Context) (*Stats, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/api/sql/stats", nil)
	if err != nil {
		return nil, err
	}
	var s Stats
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("decode stats: %w", err)
	}
	return &s, nil
}

// HostStats is returned by GET /host-stats (may be null if no data yet).
type HostStats struct {
	CPUPercent  float64 `json:"cpu_percent"`
	RAMTotal    int64   `json:"ram_total"`
	RAMUsed     int64   `json:"ram_used"`
	RAMPercent  float64 `json:"ram_percent"`
	DiskTotal   int64   `json:"disk_total"`
	DiskUsed    int64   `json:"disk_used"`
	DiskPercent float64 `json:"disk_percent"`
	Timestamp   string  `json:"timestamp"`
}

func (c *Client) GetHostStats(ctx context.Context) (*HostStats, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/api/sql/host-stats", nil)
	if err != nil {
		return nil, err
	}
	if string(data) == "null" {
		return nil, nil
	}
	var h HostStats
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("decode host stats: %w", err)
	}
	return &h, nil
}

// PGDatabase is one entry from GET /servers/{s}/pg-databases.
type PGDatabase struct {
	DatabaseName      string `json:"database_name"`
	Owner             string `json:"owner"`
	Encoding          string `json:"encoding"`
	Collation         string `json:"collation"`
	SizeBytes         int64  `json:"size_bytes"`
	ConnectionLimit   int64  `json:"connection_limit"`
	IsTemplate        bool   `json:"is_template"`
	AllowsConnections bool   `json:"allows_connections"`
}

type PGDatabasesResponse struct {
	ServerName string       `json:"server_name"`
	Databases  []PGDatabase `json:"databases"`
	UpdatedAt  *string      `json:"updated_at"`
}

func (c *Client) GetPGDatabases(ctx context.Context, serverName string) (*PGDatabasesResponse, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/api/sql/servers/"+serverName+"/pg-databases", nil)
	if err != nil {
		return nil, err
	}
	var r PGDatabasesResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("decode pg databases: %w", err)
	}
	return &r, nil
}

// PGUser is one role from GET /servers/{s}/pg-users.
type PGUser struct {
	Username           string   `json:"username"`
	IsSuperuser        bool     `json:"is_superuser"`
	CanCreateRoles     bool     `json:"can_create_roles"`
	CanCreateDB        bool     `json:"can_create_db"`
	CanLogin           bool     `json:"can_login"`
	IsReplicationUser  bool     `json:"is_replication_user"`
	PasswordValidUntil *string  `json:"password_valid_until"`
	MemberOfGroups     []string `json:"member_of_groups"`
}

type PGUsersResponse struct {
	ServerName string   `json:"server_name"`
	Users      []PGUser `json:"users"`
	UpdatedAt  *string  `json:"updated_at"`
}

func (c *Client) GetPGUsers(ctx context.Context, serverName string) (*PGUsersResponse, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/api/sql/servers/"+serverName+"/pg-users", nil)
	if err != nil {
		return nil, err
	}
	var r PGUsersResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("decode pg users: %w", err)
	}
	return &r, nil
}

// User is returned by GET /users.
type User struct {
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Picture   string  `json:"picture"`
	Role      string  `json:"role"`
	LastLogin *string `json:"last_login"`
	CreatedAt string  `json:"created_at"`
}

func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/api/sql/users", nil)
	if err != nil {
		return nil, err
	}
	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}
	return users, nil
}
