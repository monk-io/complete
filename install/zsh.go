package install

import (
	"fmt"
)

// (un)install in zsh
// basically adds/remove from .zshrc:
//
// autoload -U +X bashcompinit && bashcompinit"
// complete -C </path/to/completion/command> <command>
type zsh struct {
	rc string
}

func (z zsh) IsInstalled(cmd, bin string) bool {
	completeCmd := z.cmd(cmd, bin)
	return lineInFile(z.rc, completeCmd)
}

func (z zsh) Install(cmd, bin string) error {
	if z.IsInstalled(cmd, bin) {
		return fmt.Errorf("already installed in %s", z.rc)
	}
	var bashCompInit = "autoload -U +X bashcompinit && bashcompinit"
	if !lineInFile(z.rc, bashCompInit) {
		if err := appendFile(z.rc, "\n"+bashCompInit); err != nil {
			return err
		}
	}
	completeCmd := z.cmd(cmd, bin)
	if !lineInFile(z.rc, completeCmd) {
		if err := appendFile(z.rc, "\n"+completeCmd); err != nil {
			return err
		}
	}

	return nil
}

func (z zsh) Uninstall(cmd, bin string) error {
	if !z.IsInstalled(cmd, bin) {
		return fmt.Errorf("does not installed in %s", z.rc)
	}

	completeCmd := z.cmd(cmd, bin)
	return removeFromFile(z.rc, completeCmd)
}

func (zsh) cmd(cmd, bin string) string {
	return fmt.Sprintf("if [[ -x %s ]]; then complete -o nospace -C %s %s; fi", bin, bin, cmd)
}
