package synth

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Play toca um arquivo de áudio usando o tocador padrão do sistema
// operacional (afplay no macOS, paplay/aplay/ffplay no Linux, PowerShell
// no Windows). Requer que um desses esteja instalado.
func Play(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("afplay", path).Run()
	case "linux":
		player, err := findLinuxPlayer()
		if err != nil {
			return err
		}
		return linuxPlayerCommand(player, path).Run()
	case "windows":
		script := fmt.Sprintf(`(New-Object Media.SoundPlayer '%s').PlaySync();`, path)
		return exec.Command("powershell", "-c", script).Run()
	default:
		return fmt.Errorf("reprodução de áudio não suportada em %s", runtime.GOOS)
	}
}

func findLinuxPlayer() (string, error) {
	for _, p := range []string{"paplay", "aplay", "ffplay"} {
		if _, err := exec.LookPath(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("nenhum player de áudio encontrado (instale paplay, aplay ou ffplay)")
}

func linuxPlayerCommand(player, path string) *exec.Cmd {
	if player == "ffplay" {
		return exec.Command(player, "-nodisp", "-autoexit", "-loglevel", "quiet", path)
	}
	return exec.Command(player, path)
}
