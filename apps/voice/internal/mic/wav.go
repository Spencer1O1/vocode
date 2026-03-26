package mic

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// EncodeWavPCM16LE wraps raw PCM (s16le) into a WAV container.
func EncodeWavPCM16LE(pcm []byte, sampleRateHz int, channels int) ([]byte, error) {
	if sampleRateHz <= 0 {
		return nil, fmt.Errorf("invalid sampleRateHz %d", sampleRateHz)
	}
	if channels <= 0 {
		return nil, fmt.Errorf("invalid channels %d", channels)
	}
	if len(pcm)%2 != 0 {
		return nil, fmt.Errorf("pcm byte length must be even (s16le)")
	}

	const (
		audioFormatPCM   = 1
		bitsPerSample    = 16
		fmtChunkSize     = 16
		riffHeaderSize   = 12
		dataHeaderSize   = 8
		wavHeaderMinSize = riffHeaderSize + 8 + fmtChunkSize + dataHeaderSize
	)

	byteRate := sampleRateHz * channels * (bitsPerSample / 8)
	blockAlign := channels * (bitsPerSample / 8)

	var b bytes.Buffer
	b.Grow(wavHeaderMinSize + len(pcm))

	// RIFF header
	_, _ = b.WriteString("RIFF")
	_ = binary.Write(&b, binary.LittleEndian, uint32(36+len(pcm)))
	_, _ = b.WriteString("WAVE")

	// fmt chunk
	_, _ = b.WriteString("fmt ")
	_ = binary.Write(&b, binary.LittleEndian, uint32(fmtChunkSize))
	_ = binary.Write(&b, binary.LittleEndian, uint16(audioFormatPCM))
	_ = binary.Write(&b, binary.LittleEndian, uint16(channels))
	_ = binary.Write(&b, binary.LittleEndian, uint32(sampleRateHz))
	_ = binary.Write(&b, binary.LittleEndian, uint32(byteRate))
	_ = binary.Write(&b, binary.LittleEndian, uint16(blockAlign))
	_ = binary.Write(&b, binary.LittleEndian, uint16(bitsPerSample))

	// data chunk
	_, _ = b.WriteString("data")
	_ = binary.Write(&b, binary.LittleEndian, uint32(len(pcm)))
	_, _ = b.Write(pcm)

	return b.Bytes(), nil
}
