package utils

import (
	"bytes"
	"image"
	"image/png"
)

func ImgToBytes(img image.Image) ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
