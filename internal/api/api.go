package api

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rockpanel/rockpanel/internal/auth"
	"github.com/rockpanel/rockpanel/internal/config"
	"github.com/rockpanel/rockpanel/internal/core"
	"github.com/rockpanel/rockpanel/internal/files"
	"github.com/rockpanel/rockpanel/pkg/types"
)

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/auth/login", handleLogin).Methods("POST")
	r.HandleFunc("/api/v1/auth/logout", handleLogout).Methods("POST")
	r.HandleFunc("/api/v1/auth/me", handleGetCurrentUser).Methods("GET")

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/metrics", handleGetMetrics).Methods("GET")

	api.HandleFunc("/servers", handleListServers).Methods("GET")
	api.HandleFunc("/servers", handleCreateServer).Methods("POST")
	api.HandleFunc("/servers/{id}", handleGetServer).Methods("GET")
	api.HandleFunc("/servers/{id}", handleUpdateServer).Methods("PUT")
	api.HandleFunc("/servers/{id}", handleDeleteServer).Methods("DELETE")
	api.HandleFunc("/servers/{id}/start", handleStartServer).Methods("POST")
	api.HandleFunc("/servers/{id}/stop", handleStopServer).Methods("POST")
	api.HandleFunc("/servers/{id}/restart", handleRestartServer).Methods("POST")

	api.HandleFunc("/applications", handleListApplications).Methods("GET")
	api.HandleFunc("/applications", handleCreateApplication).Methods("POST")
	api.HandleFunc("/applications/{id}", handleGetApplication).Methods("GET")
	api.HandleFunc("/applications/{id}", handleUpdateApplication).Methods("PUT")
	api.HandleFunc("/applications/{id}", handleDeleteApplication).Methods("DELETE")
	api.HandleFunc("/applications/{id}/start", handleStartApplication).Methods("POST")
	api.HandleFunc("/applications/{id}/stop", handleStopApplication).Methods("POST")
	api.HandleFunc("/applications/{id}/restart", handleRestartApplication).Methods("POST")

	api.HandleFunc("/minecraft", handleListMinecraft).Methods("GET")
	api.HandleFunc("/minecraft", handleCreateMinecraft).Methods("POST")
	api.HandleFunc("/minecraft/{id}", handleGetMinecraft).Methods("GET")
	api.HandleFunc("/minecraft/{id}", handleUpdateMinecraft).Methods("PUT")
	api.HandleFunc("/minecraft/{id}", handleDeleteMinecraft).Methods("DELETE")
	api.HandleFunc("/minecraft/{id}/start", handleStartMinecraft).Methods("POST")
	api.HandleFunc("/minecraft/{id}/stop", handleStopMinecraft).Methods("POST")
	api.HandleFunc("/minecraft/{id}/restart", handleRestartMinecraft).Methods("POST")
	api.HandleFunc("/minecraft/{id}/command", handleMinecraftCommand).Methods("POST")

	api.HandleFunc("/docker/containers", handleListContainers).Methods("GET")
	api.HandleFunc("/docker/images", handleListImages).Methods("GET")
	api.HandleFunc("/docker/containers/{id}/start", handleStartContainer).Methods("POST")
	api.HandleFunc("/docker/containers/{id}/stop", handleStopContainer).Methods("POST")
	api.HandleFunc("/docker/containers/{id}/restart", handleRestartContainer).Methods("POST")

	api.HandleFunc("/files", handleListFiles).Methods("GET")
	api.HandleFunc("/files/upload", handleUploadFile).Methods("POST")
	api.HandleFunc("/files/{path:.*}", handleReadFile).Methods("GET")
	api.HandleFunc("/files/{path:.*}", handleWriteFile).Methods("PUT")
	api.HandleFunc("/files/{path:.*}", handleDeleteFile).Methods("DELETE")
	api.HandleFunc("/files/{path:.*}/mkdir", handleMkdir).Methods("POST")
	api.HandleFunc("/files/{path:.*}/rename", handleRenameFile).Methods("POST")
	api.HandleFunc("/files/{path:.*}/extract", handleExtractFile).Methods("POST")
	api.HandleFunc("/files/{path:.*}/compress", handleCompressFile).Methods("POST")

	api.HandleFunc("/backups", handleListBackups).Methods("GET")
	api.HandleFunc("/backups", handleCreateBackup).Methods("POST")
	api.HandleFunc("/backups/{id}", handleDeleteBackup).Methods("DELETE")
	api.HandleFunc("/backups/{id}/download", handleDownloadBackup).Methods("GET")

	api.HandleFunc("/schedules", handleListSchedules).Methods("GET")
	api.HandleFunc("/schedules", handleCreateSchedule).Methods("POST")
	api.HandleFunc("/schedules/{id}", handleUpdateSchedule).Methods("PUT")
	api.HandleFunc("/schedules/{id}", handleDeleteSchedule).Methods("DELETE")

	api.HandleFunc("/websites", handleListWebsites).Methods("GET")
	api.HandleFunc("/websites", handleCreateWebsite).Methods("POST")
	api.HandleFunc("/websites/{id}", handleUpdateWebsite).Methods("PUT")
	api.HandleFunc("/websites/{id}", handleDeleteWebsite).Methods("DELETE")

	api.HandleFunc("/databases", handleListDatabases).Methods("GET")
	api.HandleFunc("/databases", handleCreateDatabase).Methods("POST")
	api.HandleFunc("/databases/{id}", handleUpdateDatabase).Methods("PUT")
	api.HandleFunc("/databases/{id}", handleDeleteDatabase).Methods("DELETE")

	api.HandleFunc("/users", handleListUsers).Methods("GET")
	api.HandleFunc("/users", handleCreateUser).Methods("POST")
	api.HandleFunc("/users/{id}", handleUpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id}", handleDeleteUser).Methods("DELETE")

	api.HandleFunc("/tokens", handleListTokens).Methods("GET")
	api.HandleFunc("/tokens", handleCreateToken).Methods("POST")
	api.HandleFunc("/tokens/{id}", handleRevokeToken).Methods("DELETE")

	api.HandleFunc("/logs", handleGetLogs).Methods("GET")
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func parseID(r *http.Request) int64 {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	return id
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	user, err := core.GetUserByUsername(req.Username)
	if err != nil || user == nil || !auth.CheckPassword(req.Password, user.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	if err := auth.CreateSession(w, r, user); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "session failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	auth.Logout(w, r)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromSession(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}
	user.Password = ""
	writeJSON(w, http.StatusOK, user)
}

func handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, core.CollectMetrics())
}

func handleListServers(w http.ResponseWriter, r *http.Request) {
	rows, err := core.DB.Query(`SELECT id, name, command, work_dir, env_vars, user, cpu_limit, memory_limit, restart_policy, ports, status, pid, created_at, updated_at FROM servers`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var servers []types.Server
	for rows.Next() {
		var s types.Server
		rows.Scan(&s.ID, &s.Name, &s.Command, &s.WorkDir, &s.EnvVars, &s.RunUser, &s.CPULimit, &s.MemoryLimit, &s.RestartPolicy, &s.Ports, &s.Status, &s.PID, &s.CreatedAt, &s.UpdatedAt)
		servers = append(servers, s)
	}
	writeJSON(w, http.StatusOK, servers)
}

func handleCreateServer(w http.ResponseWriter, r *http.Request) {
	var s types.Server
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	now := time.Now().Unix()
	s.CreatedAt = now
	s.UpdatedAt = now
	s.Status = "stopped"
	res, err := core.DB.Exec(
		`INSERT INTO servers (name, command, work_dir, env_vars, user, cpu_limit, memory_limit, restart_policy, ports, status, pid, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.Name, s.Command, s.WorkDir, s.EnvVars, s.RunUser, s.CPULimit, s.MemoryLimit, s.RestartPolicy, s.Ports, s.Status, s.PID, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.ID, _ = res.LastInsertId()
	writeJSON(w, http.StatusCreated, s)
}

func handleGetServer(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var s types.Server
	err := core.DB.QueryRow(`SELECT id, name, command, work_dir, env_vars, user, cpu_limit, memory_limit, restart_policy, ports, status, pid, created_at, updated_at FROM servers WHERE id=?`, id).
		Scan(&s.ID, &s.Name, &s.Command, &s.WorkDir, &s.EnvVars, &s.RunUser, &s.CPULimit, &s.MemoryLimit, &s.RestartPolicy, &s.Ports, &s.Status, &s.PID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func handleUpdateServer(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var s types.Server
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.UpdatedAt = time.Now().Unix()
	_, err := core.DB.Exec(
		`UPDATE servers SET name=?, command=?, work_dir=?, env_vars=?, user=?, cpu_limit=?, memory_limit=?, restart_policy=?, ports=?, updated_at=? WHERE id=?`,
		s.Name, s.Command, s.WorkDir, s.EnvVars, s.RunUser, s.CPULimit, s.MemoryLimit, s.RestartPolicy, s.Ports, s.UpdatedAt, id,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.ID = id
	writeJSON(w, http.StatusOK, s)
}

func handleDeleteServer(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.StopServer(id)
	core.DB.Exec(`DELETE FROM servers WHERE id=?`, id)
	writeJSON(w, http.StatusNoContent, nil)
}

func handleStartServer(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var s types.Server
	err := core.DB.QueryRow(`SELECT id, name, command, work_dir, env_vars, user, cpu_limit, memory_limit, restart_policy, ports, status, pid, created_at, updated_at FROM servers WHERE id=?`, id).
		Scan(&s.ID, &s.Name, &s.Command, &s.WorkDir, &s.EnvVars, &s.RunUser, &s.CPULimit, &s.MemoryLimit, &s.RestartPolicy, &s.Ports, &s.Status, &s.PID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err := core.StartServer(&s); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func handleStopServer(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	if err := core.StopServer(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func handleRestartServer(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.StopServer(id)
	time.Sleep(time.Second)
	var s types.Server
	core.DB.QueryRow(`SELECT id, name, command, work_dir, env_vars, user, cpu_limit, memory_limit, restart_policy, ports, status, pid, created_at, updated_at FROM servers WHERE id=?`, id).
		Scan(&s.ID, &s.Name, &s.Command, &s.WorkDir, &s.EnvVars, &s.RunUser, &s.CPULimit, &s.MemoryLimit, &s.RestartPolicy, &s.Ports, &s.Status, &s.PID, &s.CreatedAt, &s.UpdatedAt)
	core.StartServer(&s)
	writeJSON(w, http.StatusOK, map[string]string{"status": "restarted"})
}

func handleListApplications(w http.ResponseWriter, r *http.Request) {
	rows, err := core.DB.Query(`SELECT id, name, git_repo, branch, work_dir, install_cmd, build_cmd, start_cmd, env_vars, port, restart_policy, status, pid, created_at, updated_at FROM applications`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var apps []types.Application
	for rows.Next() {
		var a types.Application
		rows.Scan(&a.ID, &a.Name, &a.GitRepo, &a.Branch, &a.WorkDir, &a.InstallCmd, &a.BuildCmd, &a.StartCmd, &a.EnvVars, &a.Port, &a.RestartPolicy, &a.Status, &a.PID, &a.CreatedAt, &a.UpdatedAt)
		apps = append(apps, a)
	}
	writeJSON(w, http.StatusOK, apps)
}

func handleCreateApplication(w http.ResponseWriter, r *http.Request) {
	var a types.Application
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	now := time.Now().Unix()
	a.CreatedAt = now
	a.UpdatedAt = now
	a.Status = "stopped"
	res, err := core.DB.Exec(
		`INSERT INTO applications (name, git_repo, branch, work_dir, install_cmd, build_cmd, start_cmd, env_vars, port, restart_policy, status, pid, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.Name, a.GitRepo, a.Branch, a.WorkDir, a.InstallCmd, a.BuildCmd, a.StartCmd, a.EnvVars, a.Port, a.RestartPolicy, a.Status, a.PID, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	a.ID, _ = res.LastInsertId()
	writeJSON(w, http.StatusCreated, a)
}

func handleGetApplication(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var a types.Application
	err := core.DB.QueryRow(`SELECT id, name, git_repo, branch, work_dir, install_cmd, build_cmd, start_cmd, env_vars, port, restart_policy, status, pid, created_at, updated_at FROM applications WHERE id=?`, id).
		Scan(&a.ID, &a.Name, &a.GitRepo, &a.Branch, &a.WorkDir, &a.InstallCmd, &a.BuildCmd, &a.StartCmd, &a.EnvVars, &a.Port, &a.RestartPolicy, &a.Status, &a.PID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func handleUpdateApplication(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var a types.Application
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	a.UpdatedAt = time.Now().Unix()
	core.DB.Exec(
		`UPDATE applications SET name=?, git_repo=?, branch=?, work_dir=?, install_cmd=?, build_cmd=?, start_cmd=?, env_vars=?, port=?, restart_policy=?, updated_at=? WHERE id=?`,
		a.Name, a.GitRepo, a.Branch, a.WorkDir, a.InstallCmd, a.BuildCmd, a.StartCmd, a.EnvVars, a.Port, a.RestartPolicy, a.UpdatedAt, id,
	)
	a.ID = id
	writeJSON(w, http.StatusOK, a)
}

func handleDeleteApplication(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.StopApplication(id)
	core.DB.Exec(`DELETE FROM applications WHERE id=?`, id)
	writeJSON(w, http.StatusNoContent, nil)
}

func handleStartApplication(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var a types.Application
	err := core.DB.QueryRow(`SELECT id, name, git_repo, branch, work_dir, install_cmd, build_cmd, start_cmd, env_vars, port, restart_policy, status, pid, created_at, updated_at FROM applications WHERE id=?`, id).
		Scan(&a.ID, &a.Name, &a.GitRepo, &a.Branch, &a.WorkDir, &a.InstallCmd, &a.BuildCmd, &a.StartCmd, &a.EnvVars, &a.Port, &a.RestartPolicy, &a.Status, &a.PID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err := core.StartApplication(&a); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func handleStopApplication(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.StopApplication(id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func handleRestartApplication(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.StopApplication(id)
	time.Sleep(time.Second)
	var a types.Application
	core.DB.QueryRow(`SELECT id, name, git_repo, branch, work_dir, install_cmd, build_cmd, start_cmd, env_vars, port, restart_policy, status, pid, created_at, updated_at FROM applications WHERE id=?`, id).
		Scan(&a.ID, &a.Name, &a.GitRepo, &a.Branch, &a.WorkDir, &a.InstallCmd, &a.BuildCmd, &a.StartCmd, &a.EnvVars, &a.Port, &a.RestartPolicy, &a.Status, &a.PID, &a.CreatedAt, &a.UpdatedAt)
	core.StartApplication(&a)
	writeJSON(w, http.StatusOK, map[string]string{"status": "restarted"})
}

func handleListMinecraft(w http.ResponseWriter, r *http.Request) {
	rows, err := core.DB.Query(`SELECT id, name, version, server_type, java_version, memory, port, work_dir, status, pid, created_at, updated_at FROM minecraft_servers`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var servers []types.MinecraftServer
	for rows.Next() {
		var s types.MinecraftServer
		rows.Scan(&s.ID, &s.Name, &s.Version, &s.ServerType, &s.JavaVersion, &s.Memory, &s.Port, &s.WorkDir, &s.Status, &s.PID, &s.CreatedAt, &s.UpdatedAt)
		servers = append(servers, s)
	}
	writeJSON(w, http.StatusOK, servers)
}

func handleCreateMinecraft(w http.ResponseWriter, r *http.Request) {
	var s types.MinecraftServer
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	now := time.Now().Unix()
	s.CreatedAt = now
	s.UpdatedAt = now
	s.Status = "stopped"
	if s.WorkDir == "" {
		s.WorkDir = filepath.Join(config.C.DataDir, "minecraft", s.Name)
	}
	res, err := core.DB.Exec(
		`INSERT INTO minecraft_servers (name, version, server_type, java_version, memory, port, work_dir, status, pid, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.Name, s.Version, s.ServerType, s.JavaVersion, s.Memory, s.Port, s.WorkDir, s.Status, s.PID, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.ID, _ = res.LastInsertId()
	writeJSON(w, http.StatusCreated, s)
}

func handleGetMinecraft(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var s types.MinecraftServer
	err := core.DB.QueryRow(`SELECT id, name, version, server_type, java_version, memory, port, work_dir, status, pid, created_at, updated_at FROM minecraft_servers WHERE id=?`, id).
		Scan(&s.ID, &s.Name, &s.Version, &s.ServerType, &s.JavaVersion, &s.Memory, &s.Port, &s.WorkDir, &s.Status, &s.PID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func handleUpdateMinecraft(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var s types.MinecraftServer
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.UpdatedAt = time.Now().Unix()
	core.DB.Exec(
		`UPDATE minecraft_servers SET name=?, version=?, server_type=?, java_version=?, memory=?, port=?, work_dir=?, updated_at=? WHERE id=?`,
		s.Name, s.Version, s.ServerType, s.JavaVersion, s.Memory, s.Port, s.WorkDir, s.UpdatedAt, id,
	)
	s.ID = id
	writeJSON(w, http.StatusOK, s)
}

func handleDeleteMinecraft(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.StopMinecraftServer(id)
	core.DB.Exec(`DELETE FROM minecraft_servers WHERE id=?`, id)
	writeJSON(w, http.StatusNoContent, nil)
}

func handleStartMinecraft(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var s types.MinecraftServer
	err := core.DB.QueryRow(`SELECT id, name, version, server_type, java_version, memory, port, work_dir, status, pid, created_at, updated_at FROM minecraft_servers WHERE id=?`, id).
		Scan(&s.ID, &s.Name, &s.Version, &s.ServerType, &s.JavaVersion, &s.Memory, &s.Port, &s.WorkDir, &s.Status, &s.PID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err := core.StartMinecraftServer(&s); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func handleStopMinecraft(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.StopMinecraftServer(id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func handleRestartMinecraft(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.StopMinecraftServer(id)
	time.Sleep(time.Second)
	var s types.MinecraftServer
	core.DB.QueryRow(`SELECT id, name, version, server_type, java_version, memory, port, work_dir, status, pid, created_at, updated_at FROM minecraft_servers WHERE id=?`, id).
		Scan(&s.ID, &s.Name, &s.Version, &s.ServerType, &s.JavaVersion, &s.Memory, &s.Port, &s.WorkDir, &s.Status, &s.PID, &s.CreatedAt, &s.UpdatedAt)
	core.StartMinecraftServer(&s)
	writeJSON(w, http.StatusOK, map[string]string{"status": "restarted"})
}

func handleMinecraftCommand(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var req struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := core.SendMinecraftCommand(id, req.Command); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func handleListContainers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []types.DockerContainer{})
}

func handleListImages(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []types.DockerImage{})
}

func handleStartContainer(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func handleStopContainer(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func handleRestartContainer(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "restarted"})
}

func handleListFiles(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("path")
	entries, err := files.List(dir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func handleUploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no file provided"})
		return
	}
	defer file.Close()
	uploadDir := r.FormValue("path")
	dest := uploadDir
	if dest != "" {
		dest += "/"
	}
	dest += header.Filename
	data, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if err := files.Write(dest, data); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "uploaded"})
}

func handleReadFile(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	data, err := files.Read(path)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func handleWriteFile(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := files.Write(path, data); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	if err := files.Remove(path); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func handleMkdir(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	if err := files.Mkdir(path); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "created"})
}

func handleRenameFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NewName string `json:"new_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	oldPath := mux.Vars(r)["path"]
	newPath := filepath.Dir(oldPath) + "/" + req.NewName
	if err := files.Rename(oldPath, newPath); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "renamed"})
}

func handleExtractFile(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	if err := files.Extract(path, filepath.Dir(path)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "extracted"})
}

func handleCompressFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Sources []string `json:"sources"`
		Dest    string   `json:"dest"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := files.Compress(req.Sources, req.Dest); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "compressed"})
}

func handleListBackups(w http.ResponseWriter, r *http.Request) {
	rows, err := core.DB.Query(`SELECT id, name, type, target_id, path, size, status, created_at, completed_at FROM backups`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var backups []types.Backup
	for rows.Next() {
		var b types.Backup
		rows.Scan(&b.ID, &b.Name, &b.Type, &b.TargetID, &b.Path, &b.Size, &b.Status, &b.CreatedAt, &b.CompletedAt)
		backups = append(backups, b)
	}
	writeJSON(w, http.StatusOK, backups)
}

func handleCreateBackup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		TargetID int64  `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	now := time.Now().Unix()
	backupPath := filepath.Join(config.C.DataDir, "backups", req.Name+".tar.gz")
	res, err := core.DB.Exec(
		`INSERT INTO backups (name, type, target_id, path, size, status, created_at, completed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.Name, req.Type, req.TargetID, backupPath, 0, "completed", now, now,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id, "status": "completed"})
}

func handleDeleteBackup(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.DB.Exec(`DELETE FROM backups WHERE id=?`, id)
	writeJSON(w, http.StatusNoContent, nil)
}

func handleDownloadBackup(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var b types.Backup
	err := core.DB.QueryRow(`SELECT path FROM backups WHERE id=?`, id).Scan(&b.Path)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	http.ServeFile(w, r, b.Path)
}

func handleListSchedules(w http.ResponseWriter, r *http.Request) {
	rows, err := core.DB.Query(`SELECT id, name, type, target_id, cron_expr, command, enabled, last_run, next_run, created_at, updated_at FROM schedules`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var schedules []types.Schedule
	for rows.Next() {
		var s types.Schedule
		rows.Scan(&s.ID, &s.Name, &s.Type, &s.TargetID, &s.CronExpr, &s.Command, &s.Enabled, &s.LastRun, &s.NextRun, &s.CreatedAt, &s.UpdatedAt)
		schedules = append(schedules, s)
	}
	writeJSON(w, http.StatusOK, schedules)
}

func handleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var s types.Schedule
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	now := time.Now().Unix()
	s.CreatedAt = now
	s.UpdatedAt = now
	res, err := core.DB.Exec(
		`INSERT INTO schedules (name, type, target_id, cron_expr, command, enabled, last_run, next_run, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.Name, s.Type, s.TargetID, s.CronExpr, s.Command, s.Enabled, s.LastRun, s.NextRun, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.ID, _ = res.LastInsertId()
	writeJSON(w, http.StatusCreated, s)
}

func handleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var s types.Schedule
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.UpdatedAt = time.Now().Unix()
	core.DB.Exec(
		`UPDATE schedules SET name=?, type=?, target_id=?, cron_expr=?, command=?, enabled=?, updated_at=? WHERE id=?`,
		s.Name, s.Type, s.TargetID, s.CronExpr, s.Command, s.Enabled, s.UpdatedAt, id,
	)
	s.ID = id
	writeJSON(w, http.StatusOK, s)
}

func handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.DB.Exec(`DELETE FROM schedules WHERE id=?`, id)
	writeJSON(w, http.StatusNoContent, nil)
}

func handleListWebsites(w http.ResponseWriter, r *http.Request) {
	rows, err := core.DB.Query(`SELECT id, domain, target, ssl_enabled, ssl_cert, ssl_key, created_at, updated_at FROM websites`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var websites []types.Website
	for rows.Next() {
		var w types.Website
		rows.Scan(&w.ID, &w.Domain, &w.Target, &w.SSLEnabled, &w.SSLCert, &w.SSLKey, &w.CreatedAt, &w.UpdatedAt)
		websites = append(websites, w)
	}
	writeJSON(w, http.StatusOK, websites)
}

func handleCreateWebsite(w http.ResponseWriter, r *http.Request) {
	var site types.Website
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	now := time.Now().Unix()
	site.CreatedAt = now
	site.UpdatedAt = now
	res, err := core.DB.Exec(
		`INSERT INTO websites (domain, target, ssl_enabled, ssl_cert, ssl_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		site.Domain, site.Target, site.SSLEnabled, site.SSLCert, site.SSLKey, site.CreatedAt, site.UpdatedAt,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	site.ID, _ = res.LastInsertId()
	writeJSON(w, http.StatusCreated, site)
}

func handleUpdateWebsite(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var site types.Website
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	site.UpdatedAt = time.Now().Unix()
	core.DB.Exec(
		`UPDATE websites SET domain=?, target=?, ssl_enabled=?, ssl_cert=?, ssl_key=?, updated_at=? WHERE id=?`,
		site.Domain, site.Target, site.SSLEnabled, site.SSLCert, site.SSLKey, site.UpdatedAt, id,
	)
	site.ID = id
	writeJSON(w, http.StatusOK, site)
}

func handleDeleteWebsite(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.DB.Exec(`DELETE FROM websites WHERE id=?`, id)
	writeJSON(w, http.StatusNoContent, nil)
}

func handleListDatabases(w http.ResponseWriter, r *http.Request) {
	rows, err := core.DB.Query(`SELECT id, name, type, host, port, username, created_at, updated_at FROM databases`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var dbs []types.Database
	for rows.Next() {
		var d types.Database
		rows.Scan(&d.ID, &d.Name, &d.Type, &d.Host, &d.Port, &d.Username, &d.CreatedAt, &d.UpdatedAt)
		dbs = append(dbs, d)
	}
	writeJSON(w, http.StatusOK, dbs)
}

func handleCreateDatabase(w http.ResponseWriter, r *http.Request) {
	var d types.Database
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	now := time.Now().Unix()
	d.CreatedAt = now
	d.UpdatedAt = now
	res, err := core.DB.Exec(
		`INSERT INTO databases (name, type, host, port, username, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		d.Name, d.Type, d.Host, d.Port, d.Username, d.Password, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	d.ID, _ = res.LastInsertId()
	d.Password = ""
	writeJSON(w, http.StatusCreated, d)
}

func handleUpdateDatabase(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var d types.Database
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	d.UpdatedAt = time.Now().Unix()
	core.DB.Exec(
		`UPDATE databases SET name=?, type=?, host=?, port=?, username=?, password=?, updated_at=? WHERE id=?`,
		d.Name, d.Type, d.Host, d.Port, d.Username, d.Password, d.UpdatedAt, id,
	)
	d.ID = id
	d.Password = ""
	writeJSON(w, http.StatusOK, d)
}

func handleDeleteDatabase(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.DB.Exec(`DELETE FROM databases WHERE id=?`, id)
	writeJSON(w, http.StatusNoContent, nil)
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := core.ListUsers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if req.Role == "" {
		req.Role = "user"
	}
	if err := core.CreateUser(req.Username, hash, req.Role); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	var req struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := core.UpdateUser(id, req.Username, req.Role); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.DeleteUser(id)
	writeJSON(w, http.StatusNoContent, nil)
}

func handleListTokens(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	tokens, err := core.ListAPITokens(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func handleCreateToken(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req struct {
		Name      string `json:"name"`
		ExpiresAt int64  `json:"expires_at"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	rawToken, prefix, err := auth.GenerateToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	hash, err := auth.HashPassword(rawToken)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if err := core.CreateAPIToken(req.Name, hash, prefix, user.ID, req.ExpiresAt); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{
		"token":  rawToken,
		"prefix": prefix,
		"name":   req.Name,
	})
}

func handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	core.RevokeAPIToken(id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

func handleGetLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []types.LogEntry{})
}