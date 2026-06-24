package resources

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
)

func GenerateTrayIcon() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	c1 := color.RGBA{99, 102, 241, 255}
	c2 := color.RGBA{168, 85, 247, 255}

	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			dx, dy := x-16, y-16
			dist := float64(dx*dx+dy*dy)
			if dist <= 140 {
				ratio := float64(y) / 31
				r := uint8(float64(c1.R)*(1-ratio) + float64(c2.R)*ratio)
				g := uint8(float64(c1.G)*(1-ratio) + float64(c2.G)*ratio)
				b := uint8(float64(c1.B)*(1-ratio) + float64(c2.B)*ratio)
				img.Set(x, y, color.RGBA{r, g, b, 255})
			} else if dist <= 200 {
				ratio := float64(y) / 31
				r := uint8(float64(c1.R)*(1-ratio) + float64(c2.R)*ratio)
				g := uint8(float64(c1.G)*(1-ratio) + float64(c2.G)*ratio)
				b := uint8(float64(c1.B)*(1-ratio) + float64(c2.B)*ratio)
				img.Set(x, y, color.RGBA{r, g, b, 100})
			}
		}
	}

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		return nil
	}
	pngData := pngBuf.Bytes()

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint16(0))
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(1))

	buf.WriteByte(32)
	buf.WriteByte(32)
	buf.WriteByte(0)
	buf.WriteByte(0)
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(32))
	binary.Write(&buf, binary.LittleEndian, uint32(len(pngData)))
	binary.Write(&buf, binary.LittleEndian, uint32(22))

	buf.Write(pngData)
	return buf.Bytes()
}
