package synth

import (
	"encoding/binary"
	"fmt"
	"os"
)

// WriteWAV escreve um WAV PCM 16-bit mono a partir de amostras em [-1, 1].
func WriteWAV(path string, samples []float64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	numSamples := len(samples)
	dataSize := numSamples * 2 // 16 bits = 2 bytes
	byteRate := SampleRate * 2

	header := make([]byte, 44)
	copy(header[0:4], "RIFF")
	binary.LittleEndian.PutUint32(header[4:8], uint32(36+dataSize))
	copy(header[8:12], "WAVE")
	copy(header[12:16], "fmt ")
	binary.LittleEndian.PutUint32(header[16:20], 16) // subchunk size
	binary.LittleEndian.PutUint16(header[20:22], 1)  // PCM
	binary.LittleEndian.PutUint16(header[22:24], 1)  // mono
	binary.LittleEndian.PutUint32(header[24:28], SampleRate)
	binary.LittleEndian.PutUint32(header[28:32], uint32(byteRate))
	binary.LittleEndian.PutUint16(header[32:34], 2)  // block align
	binary.LittleEndian.PutUint16(header[34:36], 16) // bits per sample
	copy(header[36:40], "data")
	binary.LittleEndian.PutUint32(header[40:44], uint32(dataSize))

	if _, err := f.Write(header); err != nil {
		return err
	}

	buf := make([]byte, dataSize)
	for i, s := range samples {
		if s > 1 {
			s = 1
		}
		if s < -1 {
			s = -1
		}
		v := int16(s * 32767)
		binary.LittleEndian.PutUint16(buf[i*2:i*2+2], uint16(v))
	}
	_, err = f.Write(buf)
	return err
}

// ReadWAV lê um arquivo WAV PCM 16-bit mono na taxa de amostragem do
// pacote (44100Hz) e retorna as amostras normalizadas em [-1, 1]. Outros
// formatos (estéreo, outra taxa, outra profundidade de bits) retornam
// erro explicando como converter.
func ReadWAV(path string) ([]float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, fmt.Errorf("%s: não é um arquivo WAV válido", path)
	}

	var channels, bitsPerSample uint16
	var sampleRate uint32
	dataOffset, dataSize := -1, 0

	pos := 12
	for pos+8 <= len(data) {
		id := string(data[pos : pos+4])
		size := int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		body := pos + 8
		switch id {
		case "fmt ":
			if body+16 > len(data) {
				return nil, fmt.Errorf("%s: chunk fmt corrompido", path)
			}
			channels = binary.LittleEndian.Uint16(data[body+2 : body+4])
			sampleRate = binary.LittleEndian.Uint32(data[body+4 : body+8])
			bitsPerSample = binary.LittleEndian.Uint16(data[body+14 : body+16])
		case "data":
			dataOffset = body
			dataSize = size
		}
		pos = body + size
		if size%2 == 1 { // chunks são alinhados a 2 bytes
			pos++
		}
	}
	if dataOffset < 0 {
		return nil, fmt.Errorf("%s: chunk 'data' não encontrado", path)
	}
	if channels != 1 || bitsPerSample != 16 || sampleRate != SampleRate {
		return nil, fmt.Errorf(
			"%s: esperado WAV mono 16-bit a %dHz, recebido %d canal(is)/%d bits/%dHz — converta com: ffmpeg -i entrada -ac 1 -ar %d -sample_fmt s16 saida.wav",
			path, SampleRate, channels, bitsPerSample, sampleRate, SampleRate)
	}
	if dataOffset+dataSize > len(data) {
		dataSize = len(data) - dataOffset
	}

	n := dataSize / 2
	samples := make([]float64, n)
	for i := range n {
		v := int16(binary.LittleEndian.Uint16(data[dataOffset+i*2 : dataOffset+i*2+2]))
		samples[i] = float64(v) / 32768
	}
	return samples, nil
}
