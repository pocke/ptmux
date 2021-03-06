package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

func TestMain(m *testing.M) {
	sessionID, err := exec.Command("tmux", "new-session", "-dP").Output()
	if err != nil {
		panic(err)
	}
	defer func() {
		exec.Command("tmux", "kill-session", string(sessionID[:len(sessionID)-1])).Run()
	}()

	err = exec.Command("tmux", "set", "-g", "base-index", "1").Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command("tmux", "set-option", "-g", "renumber-windows", "on").Run()
	if err != nil {
		panic(err)
	}

	ret := m.Run()
	os.Exit(ret)
}

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

	AssertWindowCount(t, sessionID, 1)
	RetryTest(t, 1*time.Second, 10, func() error {
		return AssertRunningCommand(t, sessionID, "1", []string{"watch"})
	})
}

func TestExecute_WithManyPanes(t *testing.T) {
	c := &Config{
		Windows: []Window{
			{
				Panes: []Pane{
					{Command: "watch ls"},
					{Command: "cat"},
					{Command: "yes"},
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

	AssertWindowCount(t, sessionID, 1)
	RetryTest(t, 1*time.Second, 10, func() error {
		return AssertRunningCommand(t, sessionID, "1", []string{"watch", "sh", "yes", "cat"})
	})
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

	AssertWindowCount(t, sessionID, 4)

	RetryTest(t, 1*time.Second, 10, func() error {
		return AssertRunningCommand(t, sessionID, "1", []string{"cat"})
	})
	RetryTest(t, 1*time.Second, 10, func() error {
		return AssertRunningCommand(t, sessionID, "2", []string{"yes"})
	})
	RetryTest(t, 1*time.Second, 10, func() error {
		return AssertRunningCommand(t, sessionID, "3", []string{"watch"})
	})
	RetryTest(t, 1*time.Second, 10, func() error {
		return AssertRunningCommand(t, sessionID, "4", []string{"sh"})
	})
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

	AssertWindowCount(t, sessionID, 1)
	RetryTest(t, 1*time.Second, 10, func() error {
		return AssertRunningCommand(t, sessionID, "1", []string{"./sh"})
	})
}

func TestExecute_WithEnv(t *testing.T) {
	c := &Config{
		Env: map[string]string{
			"v1": "foo",
			"v2": "bar",
			"v3": "baz",
			"v4": "foobar",
			"v5": "aaa",
		},
		Windows: []Window{
			{Panes: []Pane{
				{Command: "test $v1 = foo && watch ls"},
				{Command: "test $v2 = bar && cat"},
			}},
			{Panes: []Pane{
				{Command: "test $v3 = baz && yes"},
				{Command: "test $v4 = foobar && watch ls"},
				{Command: "test $v5 = aaa && cat"},
			}},
		},
		Attach: boolPtr(false),
	}

	sessionID, err := Execute(t, c)
	if err != nil {
		t.Error(err)
	}
	defer CleanSession(sessionID)

	AssertWindowCount(t, sessionID, 2)
	RetryTest(t, 1*time.Second, 10, func() error {
		return AssertRunningCommand(t, sessionID, "1", []string{"watch", "cat"})
	})
	RetryTest(t, 1*time.Second, 10, func() error {
		return AssertRunningCommand(t, sessionID, "2", []string{"yes", "cat", "watch"})
	})
}

func TestLoadConf(t *testing.T) {
	contentYAML := `root: ~/hogehoge
name: poyoyo
windows:
  - panes:
    - command: 'bin/rails s'
    - command: 'bundle exec sidekiq'
    - command: 'bin/rails c'
  - panes:
    - command: 'gvim'
    - command: 'bundle exec guard'
`
	contentJSON := `{
	"root": "~/hogehoge",
	"name": "poyoyo",
	"windows": [
		{
			"panes": [
				{
					"command": "bin/rails s"
				},
				{
					"command": "bundle exec sidekiq"
				},
				{
					"command": "bin/rails c"
				}
			]
		},
		{
			"panes": [
				{
					"command": "gvim"
				},
				{
					"command": "bundle exec guard"
				}
			]
		}
	]
}`

	expected := &Config{
		Root: "~/hogehoge",
		Name: "poyoyo",
		Windows: []Window{
			{
				Panes: []Pane{
					{Command: "bin/rails s"},
					{Command: "bundle exec sidekiq"},
					{Command: "bin/rails c"},
				},
			},
			{
				Panes: []Pane{
					{Command: "gvim"},
					{Command: "bundle exec guard"},
				},
			},
		},
	}

	testCases := []struct {
		path    string
		content string
	}{
		{
			path:    "~/.config/ptmux/testtest.yml",
			content: contentYAML,
		},
		{
			path:    "~/.config/ptmux/testtest.yaml",
			content: contentYAML,
		},
		{
			path:    "~/.config/ptmux/testtest.json",
			content: contentJSON,
		},
	}

	for _, c := range testCases {
		func() {
			t.Logf("For %s", c.path)
			path, err := homedir.Expand(c.path)
			if err != nil {
				t.Error(err)
			}

			err = ioutil.WriteFile(path, []byte(c.content), 0644)
			if err != nil {
				t.Error(err)
			}
			defer os.Remove(path)

			conf, err := LoadConf("testtest")
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(conf, expected) {
				t.Errorf("Expected %+v, but got %+v", expected, conf)
			}
		}()
	}
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

func RetryTest(t *testing.T, d time.Duration, n int, f func() error) {
	for i := 0; i < n; i++ {
		err := f()
		if err == nil {
			return
		}

		if i+1 == n {
			t.Error(err)
			return
		}
		t.Logf("Retrying... %d", i+1)
		time.Sleep(d)
	}
}

func AssertRunningCommand(t *testing.T, sessionID, windowID string, expected []string) error {
	s, err := execCommand(t, "tmux", "list-panes", "-t", sessionID+windowID, "-F", "#{pane_current_command}")
	if err != nil {
		t.Fatal(err)
	}
	cmds := strings.Split(strings.TrimSpace(s), "\n")
	if !reflect.DeepEqual(cmds, expected) {
		return errors.Errorf("Should execute %v, but got %v", expected, cmds)
	}
	return nil
}

func PrepareConfigDir() error {
	path, err := homedir.Expand("~/.config/ptmux")
	if err != nil {
		return err
	}

	return os.MkdirAll(path, 0755)
}

func init() {
	err := PrepareConfigDir()
	if err != nil {
		panic(err)
	}
}
