package images

import (
	"bytes"
	"image"
	"image/png"
)

// Converts an image type to a byte array
func ImgToBytes(img image.Image) ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
