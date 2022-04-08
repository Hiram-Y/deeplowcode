package graphics

import (
	"github.com/nfnt/resize"
	"image"
)

func ImageResize(src image.Image, w, h int) image.Image {
	return resize.Resize(uint(w), uint(h), src, resize.Lanczos3)
}
func ImageResizeSaveFile(src image.Image, width, height int, p string) error {
	dst := resize.Resize(uint(width), uint(height), src, resize.Lanczos3)
	return SaveImage(p, dst)
}
func ImageThumbnailSaveFile(src image.Image, width, height int, p string) error {
	//dst := resize.Thumbnail(uint(width), uint(height), src, resize.Lanczos3)
	//return SaveImage(p, dst)
	return nil
}
