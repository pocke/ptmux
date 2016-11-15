package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"

	yaml "gopkg.in/yaml.v2"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/shibukawa/shell"
)

func main() {
	if err := Main(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func Main(args []string) error {
	if len(args) != 2 {
		return errors.New("Please specify a profile name")
	}

	name := args[1]
	conf, err := LoadConf(name)
	if err != nil {
		return err
	}
	Exec(conf.ToShell())

	return nil
}

func LoadConf(name string) (*Config, error) {
	confPath, err := homedir.Expand(fmt.Sprintf("~/.config/ptmux/%s.yaml", name))
	if err != nil {
		return nil, errors.Wrap(err, "can't expand homedir")
	}
	if !Exists(confPath) {
		return nil, errors.Errorf("%s does not exist", confPath)
	}

	c := new(Config)
	b, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func Exec(shell string) error {
	bin, err := exec.LookPath("sh")
	if err != nil {
		return errors.Wrap(err, "cant look up `sh`")
	}
	args := []string{"sh", "-c", shell}
	env := os.Environ()
	return syscall.Exec(bin, args, env)
}

type Config struct {
	Root    string   `yaml:"root"`
	Windows []Window `yaml:"windows"`
}

func (c *Config) ToShell() string {
	res := ""
	if c.Root != "" {
		res += fmt.Sprintf("cd %s\n", c.Root)
	}
	res += "SESSION_NO=`tmux new-session -dP | cut -d : -f 1`\n\n"

	for idx, w := range c.Windows {
		res += w.ToShell(idx == 0)
	}

	res += "tmux attach-session -t $SESSION_NO\n"
	return res
}

type Window struct {
	Panes []Pane `yaml:"panes"`
}

func (w *Window) ToShell(isFirst bool) string {
	res := ""
	if isFirst {
		res += "WINDOW_NO=$SESSION_NO:1\n"
	} else {
		res += "WINDOW_NO=`tmux new-window -t $SESSION_NO -a -P | cut -d . -f 1`\n"
	}

	for idx, p := range w.Panes {
		res += p.ToShell(idx == 0)
	}

	res += "\n"

	return res
}

type Pane struct {
	Command string `yaml:"command"`
}

func (p *Pane) ToShell(isFirst bool) string {
	res := ""
	if isFirst {
		res += "PANE_NO=$WINDOW_NO.1\n"
	} else {
		res += "PANE_NO=`tmux split-window -t $WINDOW_NO -P`\n"
	}
	cmd := shell.Escape(p.Command)
	res += fmt.Sprintf("tmux send-keys -t $PANE_NO %s C-m\n", cmd)

	return res
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
