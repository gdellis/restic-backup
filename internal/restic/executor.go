package restic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound  = errors.New("restic not found")
	ErrExecFailed = errors.New("restic command failed")
)

type ResticExecutor struct {
	resticPath string
	repo       string
	password   string
	lock       sync.Mutex
}

type ResticOption func(*ResticExecutor)

func WithResticPath(path string) ResticOption {
	return func(e *ResticExecutor) {
		e.resticPath = path
	}
}

func WithRepository(repo string) ResticOption {
	return func(e *ResticExecutor) {
		e.repo = repo
	}
}

func WithPassword(password string) ResticOption {
	return func(e *ResticExecutor) {
		e.password = password
	}
}

func NewResticExecutor(opts ...ResticOption) (*ResticExecutor, error) {
	exec := &ResticExecutor{
		resticPath: "restic",
	}

	for _, opt := range opts {
		opt(exec)
	}

	if path := os.Getenv("RESTIC_CLIENT_PATH"); path != "" {
		exec.resticPath = path
	}

	if _, err := exec.cmd("--version").Output(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}

	return exec, nil
}

func (e *ResticExecutor) cmd(args ...string) *exec.Cmd {
	cmd := exec.Command(e.resticPath, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+e.password)
	if e.repo != "" {
		cmd.Env = append(cmd.Env, "RESTIC_REPOSITORY="+e.repo)
	}
	cmd.Env = append(cmd.Env, "RESTIC_KEY_HINT=")
	return cmd
}

func (e *ResticExecutor) Execute(ctx context.Context, args ...string) (string, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	cmd := e.cmd(args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+e.password)

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrExecFailed, err)
	}
	return "", nil
}

func (e *ResticExecutor) ExecuteJSON(ctx context.Context, args ...string) ([]byte, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	args = append(args, "--json")
	cmd := e.cmd(args...)
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+e.password)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%w: %s", ErrExecFailed, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("%w: %v", ErrExecFailed, err)
	}

	return output, nil
}

func (e *ResticExecutor) Init(ctx context.Context, opts ...InitOption) (*InitResult, error) {
	options := &InitOptions{}
	for _, opt := range opts {
		opt(options)
	}

	args := []string{"init", "--json"}
	if options.Password != "" {
		args = append(args, "--password-file", "-")
	}

	cmd := e.cmd(args...)
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+options.Password)
	cmd.Stdin = bytes.NewBufferString(options.Password)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result InitResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse init result: %w", err)
	}

	return &result, nil
}

type InitResult struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type InitOptions struct {
	Password string
}

type InitOption func(*InitOptions)

func InitWithPassword(password string) InitOption {
	return func(o *InitOptions) {
		o.Password = password
	}
}

type Snapshot struct {
	ID        string    `json:"id"`
	Time      time.Time `json:"time"`
	Tree      string    `json:"tree"`
	Paths     []string  `json:"paths"`
	Hostname  string    `json:"hostname"`
	Username  string    `json:"username"`
	UID       int       `json:"uid"`
	GID       int       `json:"gid"`
	Tags      []string  `json:"tags"`
	Parent    string    `json:"parent,omitempty"`
	Program   string    `json:"program,omitempty"`
	ExitStatus int      `json:"exit_status,omitempty"`
}

func (e *ResticExecutor) Snapshots(ctx context.Context, filters ...SnapshotFilter) ([]Snapshot, error) {
	args := []string{"snapshots", "--json"}

	for _, f := range filters {
		args = append(args, f()...)
	}

	output, err := e.ExecuteJSON(ctx, args...)
	if err != nil {
		return nil, err
	}

	var snapshots []Snapshot
	if err := json.Unmarshal(output, &snapshots); err != nil {
		return nil, fmt.Errorf("failed to parse snapshots: %w", err)
	}

	return snapshots, nil
}

type SnapshotFilter func() []string

func FilterByHost(host string) SnapshotFilter {
	return func() []string {
		return []string{"--host", host}
	}
}

func FilterByPath(path string) SnapshotFilter {
	return func() []string {
		return []string{"--path", path}
	}
}

func FilterByTag(tag string) SnapshotFilter {
	return func() []string {
		return []string{"--tag", tag}
	}
}

func FilterByID(id string) SnapshotFilter {
	return func() []string {
		return []string{id}
	}
}

func FilterLatest() SnapshotFilter {
	return func() []string {
		return []string{"--latest"}
	}
}

type BackupStats struct {
	FilesNew        int     `json:"files_new"`
	FilesChanged    int     `json:"files_changed"`
	FilesUnchanged int     `json:"files_unchanged"`
	DirsNew        int     `json:"dirs_new"`
	DirsChanged    int     `json:"dirs_changed"`
	DirsUnchanged  int     `json:"dirs_unchanged"`
	TotalBytes     int64   `json:"total_bytes_processed"`
	FilesProcessed int     `json:"files_processed"`
	BytesAdded     int64   `json:"bytes_added"`
	BytesTotal     int64   `json:"total_bytes"`
}

type BackupResult struct {
	SnapshotID string       `json:"snapshot_id"`
	Stats     BackupStats  `json:"stats"`
}

func (e *ResticExecutor) Backup(ctx context.Context, paths []string, opts ...BackupOption) (*BackupResult, error) {
	options := &BackupOptions{}
	for _, opt := range opts {
		opt(options)
	}

	args := []string{"backup", "--json"}
	args = append(args, paths...)

	if options.Exclude != nil {
		for _, ex := range options.Exclude {
			args = append(args, "--exclude", ex)
		}
	}

	if options.ExcludeFile != "" {
		args = append(args, "--exclude-file", options.ExcludeFile)
	}

	if options.Include != nil {
		for _, inc := range options.Include {
			args = append(args, "--include", inc)
		}
	}

	if options.Tags != nil {
		args = append(args, "--tag", strings.Join(options.Tags, ","))
	}

	if options.Host != "" {
		args = append(args, "--host", options.Host)
	}

	if options.DryRun {
		args = append(args, "--dry-run")
	}

	if options.Force {
		args = append(args, "--force")
	}

	if options.Parent != "" {
		args = append(args, "--parent", options.Parent)
	}

	if options.OneFileSystem {
		args = append(args, "--one-file-system")
	}

	if options.WithAtime {
		args = append(args, "--with-atime")
	}

	if options.Stdin {
		args = append(args, "--stdin")
		if options.StdinFilename != "" {
			args = append(args, "--stdin-filename", options.StdinFilename)
		}
	}

	output, err := e.ExecuteJSON(ctx, args...)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var result BackupResult
		if err := json.Unmarshal(line, &result); err != nil {
			continue
		}
		if result.SnapshotID != "" {
			return &result, nil
		}
	}

	return nil, fmt.Errorf("no snapshot created")
}

type BackupOptions struct {
	Exclude        []string
	ExcludeFile    string
	Include        []string
	Tags           []string
	Host           string
	DryRun         bool
	Force          bool
	Parent         string
	OneFileSystem  bool
	WithAtime      bool
	Stdin          bool
	StdinFilename  string
}

type BackupOption func(*BackupOptions)

func BackupWithExclude(patterns []string) BackupOption {
	return func(o *BackupOptions) {
		o.Exclude = patterns
	}
}

func BackupWithTags(tags []string) BackupOption {
	return func(o *BackupOptions) {
		o.Tags = tags
	}
}

func BackupWithHost(host string) BackupOption {
	return func(o *BackupOptions) {
		o.Host = host
	}
}

func BackupDryRun() BackupOption {
	return func(o *BackupOptions) {
		o.DryRun = true
	}
}

func BackupForce() BackupOption {
	return func(o *BackupOptions) {
		o.Force = true
	}
}

func BackupParent(parent string) BackupOption {
	return func(o *BackupOptions) {
		o.Parent = parent
	}
}

func BackupOneFileSystem() BackupOption {
	return func(o *BackupOptions) {
		o.OneFileSystem = true
	}
}

func BackupWithAtime() BackupOption {
	return func(o *BackupOptions) {
		o.WithAtime = true
	}
}

func (e *ResticExecutor) Restore(ctx context.Context, snapshotID string, opts ...RestoreOption) error {
	options := &RestoreOptions{}
	for _, opt := range opts {
		opt(options)
	}

	args := []string{"restore", snapshotID, "--json"}
	args = append(args, "--target", options.Target)

	if options.Hex {
		args = append(args, "--hex")
	}

	if options.Include != nil {
		for _, inc := range options.Include {
			args = append(args, "--include", inc)
		}
	}

	if options.Exclude != nil {
		for _, ex := range options.Exclude {
			args = append(args, "--exclude", ex)
		}
	}

	if options.Verify {
		args = append(args, "--verify")
	}

	_, err := e.ExecuteJSON(ctx, args...)
	return err
}

type RestoreOptions struct {
	Target   string
	Hex      bool
	Include  []string
	Exclude  []string
	Verify   bool
}

type RestoreOption func(*RestoreOptions)

func RestoreTo(target string) RestoreOption {
	return func(o *RestoreOptions) {
		o.Target = target
	}
}

func RestoreInclude(patterns []string) RestoreOption {
	return func(o *RestoreOptions) {
		o.Include = patterns
	}
}

func RestoreVerify() RestoreOption {
	return func(o *RestoreOptions) {
		o.Verify = true
	}
}

func (e *ResticExecutor) Forget(ctx context.Context, snapshotIDs []string, opts ...ForgetOption) error {
	options := &ForgetOptions{}
	for _, opt := range opts {
		opt(options)
	}

	args := []string{"forget"}
	args = append(args, snapshotIDs...)

	if options.DryRun {
		args = append(args, "--dry-run")
	}

	if options.KeepLast > 0 {
		args = append(args, "--keep-last", fmt.Sprintf("%d", options.KeepLast))
	}

	if options.KeepDaily > 0 {
		args = append(args, "--keep-daily", fmt.Sprintf("%d", options.KeepDaily))
	}

	if options.KeepWeekly > 0 {
		args = append(args, "--keep-weekly", fmt.Sprintf("%d", options.KeepWeekly))
	}

	if options.KeepMonthly > 0 {
		args = append(args, "--keep-monthly", fmt.Sprintf("%d", options.KeepMonthly))
	}

	if options.KeepYearly > 0 {
		args = append(args, "--keep-yearly", fmt.Sprintf("%d", options.KeepYearly))
	}

	if options.KeepTags != nil {
		args = append(args, "--keep-tag", strings.Join(options.KeepTags, ","))
	}

	if options.Prune {
		args = append(args, "--prune")
	}

	_, err := e.ExecuteJSON(ctx, args...)
	return err
}

type ForgetOptions struct {
	DryRun      bool
	KeepLast    int
	KeepDaily   int
	KeepWeekly  int
	KeepMonthly int
	KeepYearly  int
	KeepTags    []string
	Prune       bool
}

type ForgetOption func(*ForgetOptions)

func ForgetDryRun() ForgetOption {
	return func(o *ForgetOptions) {
		o.DryRun = true
	}
}

func ForgetKeepLast(n int) ForgetOption {
	return func(o *ForgetOptions) {
		o.KeepLast = n
	}
}

func ForgetKeepDaily(n int) ForgetOption {
	return func(o *ForgetOptions) {
		o.KeepDaily = n
	}
}

func ForgetKeepWeekly(n int) ForgetOption {
	return func(o *ForgetOptions) {
		o.KeepWeekly = n
	}
}

func ForgetKeepMonthly(n int) ForgetOption {
	return func(o *ForgetOptions) {
		o.KeepMonthly = n
	}
}

func ForgetKeepYearly(n int) ForgetOption {
	return func(o *ForgetOptions) {
		o.KeepYearly = n
	}
}

func ForgetPrune() ForgetOption {
	return func(o *ForgetOptions) {
		o.Prune = true
	}
}

type Stats struct {
	TotalFileCount int   `json:"total_file_count"`
	TotalSize     int64 `json:"total_size"`
}

func (e *ResticExecutor) Stats(ctx context.Context, snapshotID string, opts ...StatsOption) (*Stats, error) {
	options := &StatsOptions{Mode: "restore-size"}
	for _, opt := range opts {
		opt(options)
	}

	args := []string{"stats", "--json", "--mode", options.Mode}
	if snapshotID != "" {
		args = append(args, snapshotID)
	}

	output, err := e.ExecuteJSON(ctx, args...)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var stats Stats
		if err := json.Unmarshal(line, &stats); err != nil {
			continue
		}
		if stats.TotalFileCount > 0 || stats.TotalSize > 0 {
			return &stats, nil
		}
	}

	return nil, fmt.Errorf("no stats found")
}

type StatsOptions struct {
	Mode string
}

type StatsOption func(*StatsOptions)

func StatsMode(mode string) StatsOption {
	return func(o *StatsOptions) {
		o.Mode = mode
	}
}

func (e *ResticExecutor) Check(ctx context.Context, opts ...CheckOption) error {
	options := &CheckOptions{}
	for _, opt := range opts {
		opt(options)
	}

	args := []string{"check"}
	if options.ReadData {
		args = append(args, "--read-data")
	}
	if options.ReadDataSubset != "" {
		args = append(args, "--read-data-subset", options.ReadDataSubset)
	}

	_, err := e.Execute(ctx, args...)
	return err
}

type CheckOptions struct {
	ReadData         bool
	ReadDataSubset   string
}

type CheckOption func(*CheckOptions)

func CheckReadData() CheckOption {
	return func(o *CheckOptions) {
		o.ReadData = true
	}
}

func CheckReadDataSubset(subset string) CheckOption {
	return func(o *CheckOptions) {
		o.ReadDataSubset = subset
	}
}

func (e *ResticExecutor) Prune(ctx context.Context, opts ...PruneOption) error {
	options := &PruneOptions{}
	for _, opt := range opts {
		opt(options)
	}

	args := []string{"prune", "--json"}

	if options.MaxRepackSize > 0 {
		args = append(args, "--max-repack-size", fmt.Sprintf("%d", options.MaxRepackSize))
	}

	if options.ChunkerRBSize > 0 {
		args = append(args, "--chunker-rbsize", fmt.Sprintf("%d", options.ChunkerRBSize))
	}

	_, err := e.ExecuteJSON(ctx, args...)
	return err
}

type PruneOptions struct {
	MaxRepackSize  int64
	ChunkerRBSize  int64
}

type PruneOption func(*PruneOptions)

func PruneMaxRepackSize(size int64) PruneOption {
	return func(o *PruneOptions) {
		o.MaxRepackSize = size
	}
}

func (e *ResticExecutor) Unlock(ctx context.Context) error {
	_, err := e.Execute(ctx, "unlock")
	return err
}

func (e *ResticExecutor) Version(ctx context.Context) (string, error) {
	cmd := e.cmd("--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (e *ResticExecutor) CatSnapshot(ctx context.Context, snapshotID string) (*Snapshot, error) {
	output, err := e.ExecuteJSON(ctx, "cat", "snapshot", snapshotID)
	if err != nil {
		return nil, err
	}

	var snapshot Snapshot
	if err := json.Unmarshal(output, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	return &snapshot, nil
}

func (e *ResticExecutor) Diff(ctx context.Context, id1, id2 string) (string, error) {
	output, err := e.ExecuteJSON(ctx, "diff", id1, id2)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (e *ResticExecutor) Mount(ctx context.Context, mountPoint string, opts ...MountOption) error {
	options := &MountOptions{}
	for _, opt := range opts {
		opt(options)
	}

	args := []string{"mount", mountPoint}
	if options.AllSnapshots {
		args = append(args, "--snapshots")
	}

	_, err := e.Execute(ctx, args...)
	return err
}

type MountOptions struct {
	AllSnapshots bool
}

type MountOption func(*MountOptions)

func MountAllSnapshots() MountOption {
	return func(o *MountOptions) {
		o.AllSnapshots = true
	}
}
