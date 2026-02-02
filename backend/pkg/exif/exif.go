package exif

import (
	"fmt"
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

type EXIFData struct {
	// Core dimensions
	DateTime time.Time
	Width    int
	Height   int

	// Camera information
	Make  string
	Model string

	// GPS Location
	Latitude  *float64
	Longitude *float64
	Altitude  *float64

	// Camera settings
	ISO          *int
	Aperture     *float64
	ShutterSpeed string
	FocalLength  *float64

	// Orientation
	Orientation int
}

// ExtractEXIF extracts EXIF data from an image file
func ExtractEXIF(filePath string) (*EXIFData, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return nil, err
	}

	data := &EXIFData{
		Orientation: 1, // Default orientation
	}

	// Extract DateTime
	if tm, err := x.DateTime(); err == nil {
		data.DateTime = tm
	}

	// Extract dimensions - try PixelXDimension first
	if tag, err := x.Get(exif.PixelXDimension); err == nil {
		if val, err := tag.Int(0); err == nil {
			data.Width = val
		}
	}
	// Fallback to ImageWidth if PixelXDimension not found
	if data.Width == 0 {
		if tag, err := x.Get(exif.ImageWidth); err == nil {
			if val, err := tag.Int(0); err == nil {
				data.Width = val
			}
		}
	}

	// Extract height - try PixelYDimension first
	if tag, err := x.Get(exif.PixelYDimension); err == nil {
		if val, err := tag.Int(0); err == nil {
			data.Height = val
		}
	}
	// Fallback to ImageLength if PixelYDimension not found
	if data.Height == 0 {
		if tag, err := x.Get(exif.ImageLength); err == nil {
			if val, err := tag.Int(0); err == nil {
				data.Height = val
			}
		}
	}

	// Extract camera make
	if tag, err := x.Get(exif.Make); err == nil {
		if val, err := tag.StringVal(); err == nil {
			data.Make = val
		}
	}

	// Extract camera model
	if tag, err := x.Get(exif.Model); err == nil {
		if val, err := tag.StringVal(); err == nil {
			data.Model = val
		}
	}

	// Extract GPS coordinates
	if lat, lon, err := x.LatLong(); err == nil {
		data.Latitude = &lat
		data.Longitude = &lon
	}

	// Extract GPS altitude
	if tag, err := x.Get(exif.GPSAltitude); err == nil {
		if val, err := tag.Rat(0); err == nil {
			altitude := float64(val.Num().Int64()) / float64(val.Denom().Int64())

			// Check altitude reference (0 = above sea level, 1 = below)
			if refTag, err := x.Get(exif.GPSAltitudeRef); err == nil {
				if refVal, err := refTag.Int(0); err == nil && refVal == 1 {
					altitude = -altitude
				}
			}
			data.Altitude = &altitude
		}
	}

	// Extract ISO
	if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if val, err := tag.Int(0); err == nil {
			data.ISO = &val
		}
	}

	// Extract aperture (FNumber)
	if tag, err := x.Get(exif.FNumber); err == nil {
		if val, err := tag.Rat(0); err == nil {
			aperture := float64(val.Num().Int64()) / float64(val.Denom().Int64())
			data.Aperture = &aperture
		}
	}

	// Extract shutter speed (exposure time)
	if tag, err := x.Get(exif.ExposureTime); err == nil {
		if val, err := tag.Rat(0); err == nil {
			// Store as fraction string if denominator > 1, else as decimal
			if val.Denom().Int64() > 1 && val.Num().Int64() == 1 {
				data.ShutterSpeed = fmt.Sprintf("1/%d", val.Denom().Int64())
			} else {
				data.ShutterSpeed = fmt.Sprintf("%.1f", float64(val.Num().Int64())/float64(val.Denom().Int64()))
			}
		}
	}

	// Extract focal length
	if tag, err := x.Get(exif.FocalLength); err == nil {
		if val, err := tag.Rat(0); err == nil {
			focalLength := float64(val.Num().Int64()) / float64(val.Denom().Int64())
			data.FocalLength = &focalLength
		}
	}

	// Extract orientation
	if tag, err := x.Get(exif.Orientation); err == nil {
		if val, err := tag.Int(0); err == nil {
			data.Orientation = val
		}
	}

	return data, nil
}
