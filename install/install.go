// Package install provide installation functions of command completion.
package install

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-multierror"
)

var ErrNoShellsFound = errors.New("Did not find any shells to uninstall")

const (
	Bash = "bash"
	Zsh  = "zsh"
	Fish = "fish"
)

func Run(name string, uninstall, yes bool, out io.Writer, in io.Reader) {
	action := "install"
	if uninstall {
		action = "uninstall"
	}
	if !yes {
		fmt.Fprintf(out, "%s completion for %s? ", action, name)
		var answer string
		fmt.Fscanln(in, &answer)
		switch strings.ToLower(answer) {
		case "y", "yes":
		default:
			fmt.Fprintf(out, "Cancelling...\n")
			return
		}
	}
	fmt.Fprintf(out, action+"ing...\n")

	var err error
	if uninstall {
		err = Uninstall(name)
	} else {
		err = Install(name)
	}
	if err != nil {
		fmt.Fprintf(out, "%s failed: %s\n", action, err)
		os.Exit(1)
	}
}

type installer interface {
	IsInstalled(cmd, bin string) bool
	Install(cmd, bin string) error
	Uninstall(cmd, bin string) error
}

// EnsureInstall complete command given:
// cmd: is the command name
func EnsureInstall(cmd string, shells ...string) error {
	is := installers(shells...)
	if len(is) == 0 {
		return errors.New("Did not find any shells to install")
	}
	bin, err := getBinaryPath()
	if err != nil {
		return err
	}

	for _, i := range is {
		if i.IsInstalled(cmd, bin) {
			continue
		}
		errI := i.Install(cmd, bin)
		if errI != nil {
			err = multierror.Append(err, errI)
		}
	}

	return err
}

// Install complete command given:
// cmd: is the command name
func Install(cmd string) error {
	is := installers()
	if len(is) == 0 {
		return errors.New("Did not find any shells to install")
	}
	bin, err := getBinaryPath()
	if err != nil {
		return err
	}

	for _, i := range is {
		errI := i.Install(cmd, bin)
		if errI != nil {
			err = multierror.Append(err, errI)
		}
	}

	return err
}

// IsInstalled returns true if the completion
// for the given cmd is installed.
func IsInstalled(cmd string) bool {
	bin, err := getBinaryPath()
	if err != nil {
		return false
	}

	for _, i := range installers() {
		installed := i.IsInstalled(cmd, bin)
		if installed {
			return true
		}
	}

	return false
}

// Uninstall complete command given:
// cmd: is the command name
func Uninstall(cmd string) error {
	is := installers()
	if len(is) == 0 {
		return ErrNoShellsFound
	}
	bin, err := getBinaryPath()
	if err != nil {
		return err
	}

	for _, i := range is {
		errI := i.Uninstall(cmd, bin)
		if errI != nil {
			err = multierror.Append(err, errI)
		}
	}

	return err
}

func installers(shells ...string) (i []installer) {
	// The list of bash config files candidates where it is
	// possible to install the completion command.
	filterShells := len(shells) > 0
	var bashConfFiles []string
	if !filterShells || contains(shells, Bash) {
		switch runtime.GOOS {
		case "darwin":
			bashConfFiles = []string{".bash_profile"}
		default:
			bashConfFiles = []string{".bashrc", ".bash_profile", ".bash_login", ".profile"}
		}
		for _, rc := range bashConfFiles {
			if f := rcFile(rc); f != "" {
				i = append(i, bash{f})
				break
			}
		}
	}
	if !filterShells || contains(shells, Zsh) {
		if f := rcFile(".zshrc"); f != "" {
			i = append(i, zsh{f})
		}
	}

	if !filterShells || contains(shells, Fish) {
		if d := fishConfigDir(); d != "" {
			i = append(i, fish{d})
		}
	}
	return
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func fishConfigDir() string {
	configDir := filepath.Join(getConfigHomePath(), "fish")
	if configDir == "" {
		return ""
	}
	if info, err := os.Stat(configDir); err != nil || !info.IsDir() {
		return ""
	}
	return configDir
}

func getConfigHomePath() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		return filepath.Join(u.HomeDir, ".config")
	}
	return configHome
}

func getBinaryPath() (string, error) {
	bin, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(bin)
}

func rcFile(name string) string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	path := filepath.Join(u.HomeDir, name)
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}
