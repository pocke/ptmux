package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"

	yaml "gopkg.in/yaml.v2"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/ogier/pflag"
	"github.com/pkg/errors"
	"github.com/shibukawa/shell"
)

func main() {
	if err := Main(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Main(args []string) error {
	f := new(Flag)
	fs := pflag.NewFlagSet("ptmux", pflag.ContinueOnError)
	fs.BoolVarP(&f.PrintCommands, "print-commands", "p", false, "print shell commands (for debug)")
	err := fs.Parse(args[1:])
	if err != nil {
		if err == pflag.ErrHelp {
			return nil
		}
		return err
	}

	if len(fs.Args()) != 1 {
		return errors.New("Please specify a profile name")
	}
	name := fs.Arg(0)
	conf, err := LoadConf(name)
	if err != nil {
		return err
	}
	sh := conf.ToShell()

	if f.PrintCommands {
		fmt.Println(sh)
		return nil
	}

	return Exec(sh)
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
	args := []string{"bash", "-e", "-c", shell}
	env := os.Environ()
	return syscall.Exec(bin, args, env)
}

type Config struct {
	Root    string   `yaml:"root"`
	Name    string   `yaml:"name"`
	Windows []Window `yaml:"windows"`
	Attach  *bool    `yaml:"attach"`
}

func (c *Config) ToShell() string {
	res := ""
	if c.Root != "" {
		res += fmt.Sprintf("cd %s\n", c.Root)
	}
	sessionName := ""
	if c.Name != "" {
		sessionName = fmt.Sprintf("-s %s", c.Name)
	}

	res += fmt.Sprintf("SESSION_NO=`tmux new-session -dP %s`\n\n", sessionName)

	for idx, w := range c.Windows {
		res += w.ToShell(idx == 0)
	}

	if c.Attach == nil || *c.Attach {
		res += "tmux attach-session -t $SESSION_NO\n"
	} else {
		res += "echo $SESSION_NO\n"
	}
	return res
}

type Window struct {
	Panes []Pane `yaml:"panes"`
}

func (w *Window) ToShell(isFirst bool) string {
	res := ""
	if isFirst {
		res += "WINDOW_NO=$SESSION_NO\n"
	} else {
		res += "WINDOW_NO=`tmux new-window -t $SESSION_NO -a -P`\n"
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
		res += "PANE_NO=$WINDOW_NO\n"
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

type Flag struct {
	PrintCommands bool
}
