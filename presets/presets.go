// Package presets contém receitas prontas de efeitos sonoros estilo 8-bit,
// construídas em cima do pacote synth.
package presets

import "github.com/informeai/songo/synth"

// Preset gera as amostras de um efeito sonoro.
type Preset func() []float64

// All lista todos os presets disponíveis, na ordem em que devem ser
// exibidos/gerados.
var All = map[string]Preset{
	"coin":       Coin,
	"jump":       Jump,
	"laser":      Laser,
	"laser_lofi": LaserLofi,
	"explosion":  Explosion,
	"cannon":     Cannon,
	"bass":       Bass,
	"powerup":    Powerup,
	"gameover":   GameOver,
	"fm_bell":    FMBell,
}

// Names retorna os nomes dos presets em ordem estável (a de All acima).
func Names() []string {
	order := []string{
		"coin", "jump", "laser", "laser_lofi", "explosion",
		"cannon", "bass", "powerup", "gameover", "fm_bell",
	}
	return order
}

// Coin: arpejo rápido subindo, clássico "ping" de moeda.
func Coin() []float64 {
	out := synth.Concat(
		synth.Square(988, 0.05, 0.5, 0.4),  // B5
		synth.Square(1319, 0.15, 0.5, 0.4), // E6
	)
	synth.Envelope(out, 0.001, 0.02, 0.6, 0.05)
	return out
}

// Jump: sweep de frequência subindo rápido.
func Jump() []float64 {
	out := synth.Sweep(200, 800, 0.15, 0.5, 0.4)
	synth.Envelope(out, 0.001, 0.05, 0.3, 0.05)
	return out
}

// Laser: sweep descendente rápido em onda quadrada fina.
func Laser() []float64 {
	out := synth.Sweep(1200, 200, 0.2, 0.2, 0.4)
	synth.Envelope(out, 0.001, 0.02, 0.4, 0.1)
	return out
}

// LaserLofi: o mesmo laser, mas com bitcrush pra um timbre quebrado/lo-fi.
func LaserLofi() []float64 {
	return synth.Bitcrush(Laser(), 4, 6)
}

// Explosion: ruído com decay longo.
func Explosion() []float64 {
	out := synth.Noise(0.6, 0.6)
	synth.Envelope(out, 0.001, 0.5, 0.0, 0.1)
	return out
}

// Cannon: ruído + filtro passa-baixa que fecha ao longo do tempo + eco,
// bem mais denso e espacial que a explosão simples.
func Cannon() []float64 {
	filtered := synth.SweepFilter(synth.Noise(0.5, 0.8), 4000, 100)
	synth.Envelope(filtered, 0.001, 0.05, 0.4, 0.3)
	return synth.Delay(filtered, 0.08, 0.4, 3)
}

// Bass: onda triangular grave, tipo canal de baixo do NES.
func Bass() []float64 {
	out := synth.Triangle(110, 0.5, 0.6)
	synth.Envelope(out, 0.01, 0.1, 0.7, 0.2)
	return out
}

// Powerup: arpejo ascendente rápido terminando num tom sustentado.
func Powerup() []float64 {
	notes := []float64{523.25, 659.25, 783.99, 1046.5, 1318.5, 1568.0} // C5 E5 G5 C6 E6 G6
	var parts [][]float64
	for _, f := range notes {
		note := synth.Square(f, 0.06, 0.5, 0.4)
		synth.Envelope(note, 0.001, 0.01, 0.8, 0.03)
		parts = append(parts, note)
	}
	final := synth.Square(2093.0, 0.2, 0.5, 0.4) // C7 final sustentado
	synth.Envelope(final, 0.001, 0.02, 0.6, 0.15)
	parts = append(parts, final)
	return synth.Concat(parts...)
}

// GameOver: sequência descendente lenta terminando num grave sustentado.
func GameOver() []float64 {
	notes := []float64{392.0, 349.23, 311.13, 261.63} // G4 F4 D#4 C4
	var parts [][]float64
	for i, f := range notes {
		dur := 0.25
		release := 0.05
		if i == len(notes)-1 {
			dur = 0.8
			release = 0.4
		}
		note := synth.Square(f, dur, 0.5, 0.4)
		synth.Envelope(note, 0.005, 0.03, 0.7, release)
		parts = append(parts, note)
	}
	return synth.Concat(parts...)
}

// FMBell: sino metálico via FM synthesis (modIndex alto = timbre inarmônico).
func FMBell() []float64 {
	out := synth.FM(440, 233, 8, 1.2, 0.5)
	synth.Envelope(out, 0.001, 0.1, 0.3, 0.8)
	return out
}
