package insunit

import "github.com/gdey/dxf/format"

// Unit is the drawing unit for AutoCAD DesignCenter Blocks
type Unit uint8

const (
	Unitless     Unit = 0
	Inches            = 1
	Feet              = 2
	Miles             = 3
	Millimeters       = 4
	Centimeters       = 5
	Meters            = 6
	Kilometers        = 7
	Microinches       = 8
	Mils              = 9
	Yards             = 10
	Angstroms         = 11
	Nanometers        = 12
	Microns           = 13
	Decimeters        = 14
	Decameters        = 15
	Hectometers       = 16
	Gigameters        = 17
	Astronomical      = 18
	LightYears        = 19
	Parsecs           = 20
)

func (u Unit) Format(f format.Formatter) {
	f.WriteInt(70, int(u))
}

func (u Unit) String() string {
	f := format.NewASCII()
	u.Format(f)
	return f.Output()
}

type Type int8

const (
	Scientific Type = iota - 1 // We want decimal to be default of zero.
	Decimal                    // This will be zero.
	Engineering
	Architectural
	Fractional
	WindowsDesktop
)

func (t Type) Format(f format.Formatter) {
	// When defining the constants we subtracted two from the values
	// to insure that Decimal was the zero value
	f.WriteInt(70, int(t+2))
}
