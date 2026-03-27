package goarg

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

// --- Real-world workflow: git-like CLI ---

type GitCloneCmd struct {
	Repo   string `arg:"positional,required" help:"repository URL"`
	Dir    string `arg:"positional" help:"target directory"`
	Depth  int    `arg:"--depth" help:"shallow clone depth"`
	Branch string `arg:"-b,--branch" help:"branch to checkout"`
}

type GitCommitCmd struct {
	Message string `arg:"-m,--message,required" help:"commit message"`
	All     bool   `arg:"-a,--all" help:"stage all changes"`
	Amend   bool   `arg:"--amend" help:"amend previous commit"`
}

type GitPushCmd struct {
	Remote string `arg:"positional" help:"remote name"`
	Branch string `arg:"positional" help:"branch name"`
	Force  bool   `arg:"-f,--force" help:"force push"`
}

type GitArgs struct {
	Verbose bool          `arg:"-v,--verbose" help:"verbose output"`
	Clone   *GitCloneCmd  `arg:"subcommand:clone" help:"clone a repository"`
	Commit  *GitCommitCmd `arg:"subcommand:commit" help:"record changes"`
	Push    *GitPushCmd   `arg:"subcommand:push" help:"update remote refs"`
}

func TestE2E_GitClone(t *testing.T) {
	var args GitArgs
	p, err := NewParser(Config{Program: "git"}, &args)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"clone", "--depth", "1", "-b", "main", "https://github.com/user/repo.git", "myrepo"}); err != nil {
		t.Fatal(err)
	}
	if args.Clone == nil {
		t.Fatal("Clone subcommand not set")
	}
	if args.Clone.Repo != "https://github.com/user/repo.git" {
		t.Errorf("Repo = %q", args.Clone.Repo)
	}
	if args.Clone.Dir != "myrepo" {
		t.Errorf("Dir = %q", args.Clone.Dir)
	}
	if args.Clone.Depth != 1 {
		t.Errorf("Depth = %d", args.Clone.Depth)
	}
	if args.Clone.Branch != "main" {
		t.Errorf("Branch = %q", args.Clone.Branch)
	}
	if args.Commit != nil || args.Push != nil {
		t.Error("non-invoked subcommands should be nil")
	}
	names := p.SubcommandNames()
	if len(names) != 1 || names[0] != "clone" {
		t.Errorf("SubcommandNames = %v", names)
	}
}

func TestE2E_GitCommitWithInheritedVerbose(t *testing.T) {
	var args GitArgs
	p, err := NewParser(Config{Program: "git"}, &args)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"commit", "-m", "initial commit", "--all", "--verbose"}); err != nil {
		t.Fatal(err)
	}
	if !args.Verbose {
		t.Error("Verbose should be true (inherited)")
	}
	if args.Commit == nil {
		t.Fatal("Commit subcommand not set")
	}
	if args.Commit.Message != "initial commit" {
		t.Errorf("Message = %q", args.Commit.Message)
	}
	if !args.Commit.All {
		t.Error("All should be true")
	}
}

func TestE2E_GitPushForce(t *testing.T) {
	var args GitArgs
	err := ParseArgs(&args, []string{"push", "-f", "origin", "main"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Push == nil {
		t.Fatal("Push subcommand not set")
	}
	if !args.Push.Force {
		t.Error("Force should be true")
	}
	if args.Push.Remote != "origin" {
		t.Errorf("Remote = %q", args.Push.Remote)
	}
	if args.Push.Branch != "main" {
		t.Errorf("Branch = %q", args.Push.Branch)
	}
}

func TestE2E_GitHelp(t *testing.T) {
	var args GitArgs
	p, err := NewParser(Config{Program: "git"}, &args)
	if err != nil {
		t.Fatal(err)
	}
	parseErr := p.Parse([]string{"--help"})
	if !errors.Is(parseErr, ErrHelp) {
		t.Fatalf("expected ErrHelp, got %v", parseErr)
	}
	var buf bytes.Buffer
	p.WriteHelp(&buf)
	help := buf.String()
	for _, want := range []string{"clone", "commit", "push", "verbose", "Usage: git"} {
		if !strings.Contains(help, want) {
			t.Errorf("help missing %q", want)
		}
	}
}

func TestE2E_GitNoSubcommand(t *testing.T) {
	var args GitArgs
	err := ParseArgs(&args, []string{"--verbose"})
	if err != nil {
		t.Fatal(err)
	}
	if !args.Verbose {
		t.Error("Verbose should be true")
	}
	if args.Clone != nil || args.Commit != nil || args.Push != nil {
		t.Error("all subcommands should be nil")
	}
}

// --- Real-world workflow: Docker-like nested subcommands ---

type DockerRunCmd struct {
	Image string   `arg:"positional,required" help:"image name"`
	Env   []string `arg:"-e,--env" help:"environment variables"`
	Port  []string `arg:"-p,--port" help:"port mappings"`
}

type DockerBuildCmd struct {
	Tag     string `arg:"-t,--tag" help:"image tag"`
	Context string `arg:"positional" help:"build context"`
}

type DockerArgs struct {
	Debug bool            `arg:"-D,--debug" help:"debug mode"`
	Run   *DockerRunCmd   `arg:"subcommand:run" help:"run a container"`
	Build *DockerBuildCmd `arg:"subcommand:build" help:"build an image"`
}

func TestE2E_DockerRun(t *testing.T) {
	var args DockerArgs
	err := ParseArgs(&args, []string{"run", "-e", "FOO=bar", "-e", "BAZ=qux", "-p", "8080:80", "nginx:latest"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Run == nil {
		t.Fatal("Run not set")
	}
	if args.Run.Image != "nginx:latest" {
		t.Errorf("Image = %q", args.Run.Image)
	}
	if len(args.Run.Env) != 2 {
		t.Errorf("Env = %v", args.Run.Env)
	}
	if len(args.Run.Port) != 1 || args.Run.Port[0] != "8080:80" {
		t.Errorf("Port = %v", args.Run.Port)
	}
}

// --- Real-world: env-only + defaults + required ---

type ServerConfig struct {
	Port     int    `arg:"-p,--port" default:"8080" help:"listen port"`
	Host     string `arg:"--host" default:"0.0.0.0" help:"bind address"`
	DBUrl    string `arg:"env:DATABASE_URL,required" help:"database connection string"`
	LogLevel string `arg:"--log-level" default:"info" help:"log level"`
}

func TestE2E_ServerConfigFromEnv(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/mydb")
	defer os.Unsetenv("DATABASE_URL")

	var cfg ServerConfig
	err := ParseArgs(&cfg, []string{"--port", "9090"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 9090 {
		t.Errorf("Port = %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host = %q (default should apply)", cfg.Host)
	}
	if cfg.DBUrl != "postgres://localhost/mydb" {
		t.Errorf("DBUrl = %q", cfg.DBUrl)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q", cfg.LogLevel)
	}
}

// --- Embedded struct with subcommands ---

type GlobalOpts struct {
	Verbose bool   `arg:"-v,--verbose" help:"verbose"`
	Config  string `arg:"-c,--config" help:"config file"`
}

type DeployCmd struct {
	Target string `arg:"positional,required" help:"deploy target"`
	DryRun bool   `arg:"--dry-run" help:"dry run mode"`
}

type AppArgs struct {
	GlobalOpts
	Deploy *DeployCmd `arg:"subcommand:deploy" help:"deploy application"`
}

func TestE2E_EmbeddedWithSubcommand(t *testing.T) {
	var args AppArgs
	err := ParseArgs(&args, []string{"--config", "prod.yml", "deploy", "--dry-run", "--verbose", "production"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Config != "prod.yml" {
		t.Errorf("Config = %q", args.Config)
	}
	if !args.Verbose {
		t.Error("Verbose should be true (inherited)")
	}
	if args.Deploy == nil {
		t.Fatal("Deploy not set")
	}
	if !args.Deploy.DryRun {
		t.Error("DryRun should be true")
	}
	if args.Deploy.Target != "production" {
		t.Errorf("Target = %q", args.Deploy.Target)
	}
}
