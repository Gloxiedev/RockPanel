package core

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rockpanel/rockpanel/pkg/types"
)

type ManagedProcess struct {
	ID          int64
	Name        string
	Cmd         *exec.Cmd
	PID         int
	StartedAt   time.Time
	stdinWriter io.WriteCloser
	mu          sync.Mutex
}

var (
	activeProcesses = make(map[int64]*ManagedProcess)
	processMu       sync.RWMutex
)

func StartServer(s *types.Server) error {
	processMu.Lock()
	defer processMu.Unlock()

	if p, ok := activeProcesses[s.ID]; ok {
		if p.Cmd != nil && p.Cmd.Process != nil {
			return nil
		}
		delete(activeProcesses, s.ID)
	}

	args := splitCommand(s.Command)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = s.WorkDir
	cmd.Env = mergeEnv(s.EnvVars)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	p := &ManagedProcess{
		ID:        s.ID,
		Name:      s.Name,
		Cmd:       cmd,
		PID:       cmd.Process.Pid,
		StartedAt: time.Now(),
	}
	activeProcesses[s.ID] = p

	go watchProcess(s.ID, cmd, "servers", s.RestartPolicy, func() {
		StartServer(s)
	})

	DB.Exec(`UPDATE servers SET status='running', pid=?, updated_at=? WHERE id=?`,
		cmd.Process.Pid, time.Now().Unix(), s.ID)
	return nil
}

func StopServer(id int64) error {
	return stopProcess(id, "servers")
}

func StartApplication(a *types.Application) error {
	processMu.Lock()
	defer processMu.Unlock()

	if p, ok := activeProcesses[a.ID]; ok {
		if p.Cmd != nil && p.Cmd.Process != nil {
			return nil
		}
		delete(activeProcesses, a.ID)
	}

	args := splitCommand(a.StartCmd)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = a.WorkDir
	cmd.Env = mergeEnv(a.EnvVars)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return err
	}

	p := &ManagedProcess{
		ID:        a.ID,
		Name:      a.Name,
		Cmd:       cmd,
		PID:       cmd.Process.Pid,
		StartedAt: time.Now(),
	}
	activeProcesses[a.ID] = p

	go watchProcess(a.ID, cmd, "applications", a.RestartPolicy, func() {
		StartApplication(a)
	})

	DB.Exec(`UPDATE applications SET status='running', pid=?, updated_at=? WHERE id=?`,
		cmd.Process.Pid, time.Now().Unix(), a.ID)
	return nil
}

func StopApplication(id int64) error {
	return stopProcess(id, "applications")
}

func StartMinecraftServer(mc *types.MinecraftServer) error {
	processMu.Lock()
	defer processMu.Unlock()

	if p, ok := activeProcesses[mc.ID]; ok {
		if p.Cmd != nil && p.Cmd.Process != nil {
			return nil
		}
		delete(activeProcesses, mc.ID)
	}

	javaBin := resolveJavaBinary(mc.JavaVersion)
	jarPath := findServerJar(mc.WorkDir, mc.ServerType)
	if jarPath == "" {
		return os.ErrNotExist
	}

	args := []string{
		"-Xms" + strconv.Itoa(mc.Memory) + "M",
		"-Xmx" + strconv.Itoa(mc.Memory) + "M",
		"-jar", jarPath,
		"nogui",
	}

	cmd := exec.Command(javaBin, args...)
	cmd.Dir = mc.WorkDir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		stdinPipe.Close()
		return err
	}

	w := stdinPipe
	p := &ManagedProcess{
		ID:          mc.ID,
		Name:        mc.Name,
		Cmd:         cmd,
		PID:         cmd.Process.Pid,
		StartedAt:   time.Now(),
		stdinWriter: w,
	}
	activeProcesses[mc.ID] = p

	go watchProcess(mc.ID, cmd, "minecraft_servers", "always", nil)

	DB.Exec(`UPDATE minecraft_servers SET status='running', pid=?, updated_at=? WHERE id=?`,
		cmd.Process.Pid, time.Now().Unix(), mc.ID)
	return nil
}

func StopMinecraftServer(id int64) error {
	return stopProcess(id, "minecraft_servers")
}

func SendMinecraftCommand(id int64, command string) error {
	processMu.RLock()
	p, ok := activeProcesses[id]
	processMu.RUnlock()
	if !ok || p.Cmd == nil || p.Cmd.Process == nil {
		return os.ErrProcessDone
	}
	if p.stdinWriter == nil {
		return os.ErrInvalid
	}
	_, err := p.stdinWriter.Write([]byte(command + "\n"))
	return err
}

func GetProcessStatus(id int64) (string, int) {
	processMu.RLock()
	defer processMu.RUnlock()
	p, ok := activeProcesses[id]
	if !ok || p.Cmd == nil || p.Cmd.Process == nil {
		return "stopped", 0
	}
	return "running", p.PID
}

func stopProcess(id int64, table string) error {
	processMu.Lock()
	p, ok := activeProcesses[id]
	if ok {
		delete(activeProcesses, id)
	}
	processMu.Unlock()

	if !ok || p == nil || p.Cmd == nil || p.Cmd.Process == nil {
		DB.Exec(`UPDATE `+table+` SET status='stopped', pid=0, updated_at=? WHERE id=?`,
			time.Now().Unix(), id)
		return nil
	}

	p.mu.Lock()
	p.Cmd.Process.Signal(syscall.SIGTERM)
	p.mu.Unlock()

	go func() {
		time.Sleep(10 * time.Second)
		processMu.RLock()
		pp, stillTracked := activeProcesses[id]
		processMu.RUnlock()
		if !stillTracked {
			return
		}
		pp.mu.Lock()
		if pp.Cmd != nil && pp.Cmd.Process != nil {
			pp.Cmd.Process.Kill()
		}
		pp.mu.Unlock()
	}()

	DB.Exec(`UPDATE `+table+` SET status='stopped', pid=0, updated_at=? WHERE id=?`,
		time.Now().Unix(), id)
	return nil
}

func watchProcess(id int64, cmd *exec.Cmd, table, restartPolicy string, onRestart func()) {
	cmd.Wait()
	processMu.Lock()
	delete(activeProcesses, id)
	processMu.Unlock()

	DB.Exec(`UPDATE `+table+` SET status='stopped', pid=0, updated_at=? WHERE id=?`,
		time.Now().Unix(), id)

	if onRestart != nil && (restartPolicy == "always" || (restartPolicy == "on-failure" && cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0)) {
		time.Sleep(3 * time.Second)
		onRestart()
	}
}

func splitCommand(cmd string) []string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return []string{}
	}

	var result []string
	var current strings.Builder
	inSingle := false
	inDouble := false

	for _, r := range cmd {
		switch {
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case r == ' ' && !inSingle && !inDouble:
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}

func mergeEnv(envVars string) []string {
	env := os.Environ()
	if envVars == "" {
		return env
	}
	for _, line := range strings.Split(envVars, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && strings.Contains(line, "=") {
			env = append(env, line)
		}
	}
	return env
}

func resolveJavaBinary(javaVersion string) string {
	if javaVersion == "" {
		javaVersion = "java17"
	}

	knownPaths := []string{
		filepath.Join("/usr/lib/jvm", javaVersion, "bin", "java"),
		filepath.Join("/usr/lib/jvm", strings.ReplaceAll(javaVersion, "java", "java-"), "bin", "java"),
		"/usr/bin/java",
		"/usr/local/bin/java",
	}

	for _, p := range knownPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "java"
}

func findServerJar(workDir, serverType string) string {
	candidates := []string{
		"server.jar",
		serverType + ".jar",
		"paper.jar",
		"purpur.jar",
		"fabric-server-launch.jar",
		"forge-universal.jar",
	}

	for _, c := range candidates {
		p := filepath.Join(workDir, c)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	entries, err := os.ReadDir(workDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.HasSuffix(name, ".jar") && name != "libraries" {
			return filepath.Join(workDir, name)
		}
	}
	return ""
}