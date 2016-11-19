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

	sessionID, err := Execute(c)
	if err != nil {
		t.Error(err)
	}
	defer CleanSession(sessionID)

	s, err := execCommand("tmux", "list-window", "-t", sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if cnt := strings.Count(strings.TrimSpace(s), "\n") + 1; cnt != 1 {
		t.Errorf("Window count should be 1, but got %d", cnt)
	}

	time.Sleep(300 * time.Millisecond)
	AssertRunningCommand(t, sessionID, "1", []string{"watch"})
}

func TestExecute_WithManyPanes(t *testing.T) {
	c := &Config{
		Windows: []Window{
			{
				Panes: []Pane{
					{Command: "watch ls"},
					{Command: "nano"},
					{Command: "yes"},
				},
			},
		},
		Attach: boolPtr(false),
	}

	sessionID, err := Execute(c)
	if err != nil {
		t.Error(err)
	}
	defer CleanSession(sessionID)

	s, err := execCommand("tmux", "list-window", "-t", sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if cnt := strings.Count(strings.TrimSpace(s), "\n") + 1; cnt != 1 {
		t.Errorf("Window count should be 1, but got %d", cnt)
	}

	time.Sleep(500 * time.Millisecond)
	AssertRunningCommand(t, sessionID, "1", []string{"watch", "nano", "yes"})
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

func Execute(c *Config) (string, error) {
	sh := c.ToShell()
	s, err := execCommand("sh", "-c", sh)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

func CleanSession(sessionID string) error {
	return exec.Command("tmux", "kill-session", "-t", sessionID).Run()
}

func execCommand(c string, args ...string) (string, error) {
	stderr := bytes.NewBuffer([]byte{})
	cmd := exec.Command(c, args...)
	cmd.Stderr = stderr
	b, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, stderr.String())
	}
	return string(b), nil
}

func AssertRunningCommand(t *testing.T, sessionID, windowID string, expected []string) {
	s, err := execCommand("tmux", "list-panes", "-t", sessionID+windowID, "-F", "#{pane_current_command}")
	if err != nil {
		t.Fatal(err)
	}
	cmds := strings.Split(strings.TrimSpace(s), "\n")
	if !reflect.DeepEqual(cmds, expected) {
		t.Errorf("Should execute %v, but got %v", expected, cmds)
	}
}
