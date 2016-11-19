package main

import (
	"bytes"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestExecute_WithSingleWindow(t *testing.T) {
	c := &Config{
		Windows: []Window{
			{
				Panes: []Pane{
					{Command: "watch ls"},
				},
			},
		},
		Attach: boolPtr(false),
	}

	sessionID, err := Execute(t, c)
	if err != nil {
		t.Error(err)
	}
	defer CleanSession(sessionID)

	time.Sleep(1 * time.Second)
	AssertWindowCount(t, sessionID, 1)
	AssertRunningCommand(t, sessionID, "1", []string{"watch"})
}

func TestExecute_WithManyPanes(t *testing.T) {
	c := &Config{
		Windows: []Window{
			{
				Panes: []Pane{
					{Command: "watch ls"},
					{Command: "cat"},
					{Command: "yes"},
				},
			},
		},
		Attach: boolPtr(false),
	}

	sessionID, err := Execute(t, c)
	if err != nil {
		t.Error(err)
	}
	defer CleanSession(sessionID)

	time.Sleep(1 * time.Second)
	AssertWindowCount(t, sessionID, 1)
	AssertRunningCommand(t, sessionID, "1", []string{"watch", "cat", "yes"})
}

func TestExecute_WithManyWindows(t *testing.T) {
	c := &Config{
		Windows: []Window{
			{
				Panes: []Pane{
					{Command: "cat"},
				},
			},
			{
				Panes: []Pane{
					{Command: "yes"},
				},
			},
			{
				Panes: []Pane{
					{Command: "watch ls"},
				},
			},
			{
				Panes: []Pane{
					{Command: "sh"},
				},
			},
		},
		Attach: boolPtr(false),
	}

	sessionID, err := Execute(t, c)
	if err != nil {
		t.Error(err)
	}
	defer CleanSession(sessionID)

	time.Sleep(1 * time.Second)
	AssertWindowCount(t, sessionID, 4)
	AssertRunningCommand(t, sessionID, "1", []string{"cat"})
	AssertRunningCommand(t, sessionID, "2", []string{"yes"})
	AssertRunningCommand(t, sessionID, "3", []string{"watch"})
	AssertRunningCommand(t, sessionID, "4", []string{"sh"})
}

func TestExecute_WithSessionName(t *testing.T) {
	c := &Config{
		Windows: []Window{
			{Panes: []Pane{{Command: "yes"}}},
		},
		Name:   "testtest",
		Attach: boolPtr(false),
	}

	sessionID, err := Execute(t, c)
	if err != nil {
		t.Error(err)
	}
	defer CleanSession(sessionID)

	AssertWindowCount(t, sessionID, 1)
	if sessionID != "testtest:" {
		t.Errorf("session id should be testtest:, but got %s", sessionID)
	}
}

func TestExecute_WithSessionRoot(t *testing.T) {
	c := &Config{
		Windows: []Window{
			{Panes: []Pane{{Command: "./sh"}}},
		},
		Root:   "/bin/",
		Attach: boolPtr(false),
	}

	sessionID, err := Execute(t, c)
	if err != nil {
		t.Error(err)
	}
	defer CleanSession(sessionID)

	time.Sleep(1 * time.Second)
	AssertWindowCount(t, sessionID, 1)
	AssertRunningCommand(t, sessionID, "1", []string{"./sh"})
}

func TestConfigToShell_WhenAttachIsNil(t *testing.T) {
	c := &Config{}

	sh := c.ToShell()
	if !strings.Contains(sh, "tmux attach-session -t $SESSION_NO") {
		t.Errorf("Can't find attach-session. code: %s", sh)
	}
}

func TestConfigToShell_WhenAttachIsTrue(t *testing.T) {
	c := &Config{
		Attach: boolPtr(true),
	}

	sh := c.ToShell()
	if !strings.Contains(sh, "tmux attach-session -t $SESSION_NO") {
		t.Errorf("Can't find attach-session. code: %s", sh)
	}
}

func TestConfigToShell_WhenAttachIsFalse(t *testing.T) {
	c := &Config{
		Attach: boolPtr(false),
	}

	sh := c.ToShell()
	if strings.Contains(sh, "tmux attach-session -t $SESSION_NO") {
		t.Errorf("attach-session is found. code:\n%s", sh)
	}
	if !strings.Contains(sh, "echo $SESSION_NO") {
		t.Errorf("Can't find printing SESSION_NO. code:\n%s", sh)
	}
}

// ---------------------------------------------- test helper
func boolPtr(b bool) *bool {
	return &b
}

func Execute(t *testing.T, c *Config) (string, error) {
	sh := c.ToShell()
	s, err := execCommand(t, "bash", "-xe", "-c", sh)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

func CleanSession(sessionID string) error {
	return exec.Command("tmux", "kill-session", "-t", sessionID).Run()
}

func execCommand(t *testing.T, c string, args ...string) (string, error) {
	stderr := bytes.NewBuffer([]byte{})
	cmd := exec.Command(c, args...)
	cmd.Stderr = stderr
	t.Log(strings.Join(cmd.Args, " "))
	b, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, stderr.String())
	}
	return string(b), nil
}

func AssertWindowCount(t *testing.T, sessionID string, expected int) {
	s, err := execCommand(t, "tmux", "list-window", "-t", sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if cnt := strings.Count(strings.TrimSpace(s), "\n") + 1; cnt != expected {
		t.Errorf("Window count should be %d, but got %d", expected, cnt)
	}
}

func AssertRunningCommand(t *testing.T, sessionID, windowID string, expected []string) {
	s, err := execCommand(t, "tmux", "list-panes", "-t", sessionID+windowID, "-F", "#{pane_current_command}")
	if err != nil {
		t.Fatal(err)
	}
	cmds := strings.Split(strings.TrimSpace(s), "\n")
	if !reflect.DeepEqual(cmds, expected) {
		t.Errorf("Should execute %v, but got %v", expected, cmds)
	}
}
