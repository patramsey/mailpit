package apiv1

import (
	"bufio"
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"net/http"
	"strings"

	"github.com/axllent/mailpit/logger"
	"github.com/axllent/mailpit/storage"
	"github.com/disintegration/imaging"
	"github.com/gorilla/mux"
	"github.com/jhillyerd/enmime"
)

var (
	thumbWidth  = 180
	thumbHeight = 120
)

// Thumbnail returns a thumbnail image for an attachment (images only)
func Thumbnail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["id"]
	partID := vars["partID"]

	a, err := storage.GetAttachmentPart(id, partID)
	if err != nil {
		httpError(w, err.Error())
		return
	}

	fileName := a.FileName
	if fileName == "" {
		fileName = a.ContentID
	}

	if !strings.HasPrefix(a.ContentType, "image/") {
		blankImage(a, w)
		return
	}

	buf := bytes.NewBuffer(a.Content)

	img, err := imaging.Decode(buf)
	if err != nil {
		// it's not an image, return default
		logger.Log().Warning(err)
		blankImage(a, w)
		return
	}

	var b bytes.Buffer
	foo := bufio.NewWriter(&b)

	var dstImageFill *image.NRGBA

	if img.Bounds().Dx() < thumbWidth || img.Bounds().Dy() < thumbHeight {
		dstImageFill = imaging.Fit(img, thumbWidth, thumbHeight, imaging.Lanczos)
	} else {
		dstImageFill = imaging.Fill(img, thumbWidth, thumbHeight, imaging.Center, imaging.Lanczos)
	}
	// create white image and paste image over the top
	// preventing black backgrounds for transparent GIF/PNG images
	dst := imaging.New(thumbWidth, thumbHeight, color.White)
	// paste the original over the top
	dst = imaging.OverlayCenter(dst, dstImageFill, 1.0)

	if err := jpeg.Encode(foo, dst, &jpeg.Options{Quality: 70}); err != nil {
		logger.Log().Warning(err)
		blankImage(a, w)
		return
	}

	w.Header().Add("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", "filename=\""+fileName+"\"")
	_, _ = w.Write(b.Bytes())
}

// Return a blank image instead of an error when file or image not supported
func blankImage(a *enmime.Part, w http.ResponseWriter) {
	rect := image.Rect(0, 0, thumbWidth, thumbHeight)
	img := image.NewRGBA(rect)
	background := color.RGBA{255, 255, 255, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{background}, image.ZP, draw.Src)
	var b bytes.Buffer
	foo := bufio.NewWriter(&b)
	dstImageFill := imaging.Fill(img, thumbWidth, thumbHeight, imaging.Center, imaging.Lanczos)

	if err := jpeg.Encode(foo, dstImageFill, &jpeg.Options{Quality: 70}); err != nil {
		logger.Log().Warning(err)
	}

	fileName := a.FileName
	if fileName == "" {
		fileName = a.ContentID
	}

	w.Header().Add("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", "filename=\""+fileName+"\"")
	_, _ = w.Write(b.Bytes())
}