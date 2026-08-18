package types

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	Role      string `json:"role"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type Server struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Command       string `json:"command"`
	WorkDir       string `json:"work_dir"`
	EnvVars       string `json:"env_vars"`
	RunUser       string `json:"user"`
	CPULimit      int    `json:"cpu_limit"`
	MemoryLimit   int64  `json:"memory_limit"`
	RestartPolicy string `json:"restart_policy"`
	Ports         string `json:"ports"`
	Status        string `json:"status"`
	PID           int    `json:"pid"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
}

type Application struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	GitRepo       string `json:"git_repo"`
	Branch        string `json:"branch"`
	WorkDir       string `json:"work_dir"`
	InstallCmd    string `json:"install_cmd"`
	BuildCmd      string `json:"build_cmd"`
	StartCmd      string `json:"start_cmd"`
	EnvVars       string `json:"env_vars"`
	Port          int    `json:"port"`
	RestartPolicy string `json:"restart_policy"`
	Status        string `json:"status"`
	PID           int    `json:"pid"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
}

type MinecraftServer struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	ServerType  string `json:"server_type"`
	JavaVersion string `json:"java_version"`
	Memory      int    `json:"memory"`
	Port        int    `json:"port"`
	WorkDir     string `json:"work_dir"`
	Status      string `json:"status"`
	PID         int    `json:"pid"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type DockerContainer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	Ports   string `json:"ports"`
	State   string `json:"state"`
	Created int64  `json:"created"`
}

type DockerImage struct {
	ID       string   `json:"id"`
	RepoTags []string `json:"repo_tags"`
	Size     int64    `json:"size"`
	Created  int64    `json:"created"`
}

type Backup struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	TargetID    int64  `json:"target_id"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	CompletedAt int64  `json:"completed_at"`
}

type Schedule struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	TargetID  int64  `json:"target_id"`
	CronExpr  string `json:"cron_expr"`
	Command   string `json:"command"`
	Enabled   bool   `json:"enabled"`
	LastRun   int64  `json:"last_run"`
	NextRun   int64  `json:"next_run"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type Website struct {
	ID         int64  `json:"id"`
	Domain     string `json:"domain"`
	Target     string `json:"target"`
	SSLEnabled bool   `json:"ssl_enabled"`
	SSLCert    string `json:"ssl_cert"`
	SSLKey     string `json:"ssl_key"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

type Database struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type APIToken struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	TokenHash string `json:"-"`
	Prefix    string `json:"prefix"`
	UserID    int64  `json:"user_id"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
	LastUsed  int64  `json:"last_used"`
}

type SystemMetrics struct {
	CPU    float64  `json:"cpu"`
	RAM    RAMInfo  `json:"ram"`
	Disk   DiskInfo `json:"disk"`
	Net    NetInfo  `json:"network"`
	Load   LoadInfo `json:"load"`
	Uptime int64    `json:"uptime"`
}

type RAMInfo struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
	Free  int64 `json:"free"`
}

type DiskInfo struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
	Free  int64 `json:"free"`
}

type NetInfo struct {
	RxBytes int64 `json:"rx_bytes"`
	TxBytes int64 `json:"tx_bytes"`
	RxSpeed int64 `json:"rx_speed"`
	TxSpeed int64 `json:"tx_speed"`
}

type LoadInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

type ProcessInfo struct {
	PID    int     `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu"`
	Memory int64   `json:"memory"`
	Status string  `json:"status"`
}

type FileEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
	Mode    string `json:"mode"`
}

type LogEntry struct {
	Timestamp int64  `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source"`
}