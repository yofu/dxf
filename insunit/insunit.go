package insunit

import (
	"strings"

	"github.com/yofu/dxf/format"
)

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
	switch u {
	case Unitless:
		return "none"
	case Inches:
		return "inches"
	case Feet:
		return "feet"
	case Miles:
		return "miles"
	case Millimeters:
		return "millimeters"
	case Centimeters:
		return "centimeters"
	case Meters:
		return "meters"
	case Kilometers:
		return "kilometers"
	case Microinches:
		return "microinches"
	case Mils:
		return "mils"
	case Yards:
		return "yards"
	case Angstroms:
		return "angstroms"
	case Nanometers:
		return "nanometers"
	case Microns:
		return "microns"
	case Decimeters:
		return "decimeters"
	case Decameters:
		return "decameters"
	case Hectometers:
		return "hectometers"
	case Gigameters:
		return "gigameters"
	case Astronomical:
		return "astronomical"
	case LightYears:
		return "light years"
	case Parsecs:
		return "parsecs"
	default:
		return "unknown"
	}
}

var str2Unit = map[string]Unit{
	"none":         Unitless,
	"unitless":     Unitless,
	"inches":       Inches,
	"feet":         Feet,
	"miles":        Miles,
	"millimeters":  Millimeters,
	"centimeters":  Centimeters,
	"meters":       Meters,
	"kilometers":   Kilometers,
	"microinches":  Microinches,
	"mils":         Mils,
	"yards":        Yards,
	"angstroms":    Angstroms,
	"nanometers":   Nanometers,
	"microns":      Microns,
	"decimeters":   Decimeters,
	"decameters":   Decameters,
	"hectometers":  Hectometers,
	"gigameters":   Gigameters,
	"astronomical": Astronomical,
	"light years":  LightYears,
	"lightyears":   LightYears,
	"light-years":  LightYears,
	"parsecs":      Parsecs,
}

func UnitFromString(str string) (Unit, bool) {
	u, ok := str2Unit[strings.TrimSpace(strings.ToLower(str))]
	return u, ok
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

func (t Type) String() string {
	switch t {
	case Scientific:
		return "scientific"
	case Decimal:
		return "decimal"
	case Engineering:
		return "engineering:"
	case Architectural:
		return "architectural:"
	case Fractional:
		return "fractional:"
	case WindowsDesktop:
		return "windows desktop"
	default:
		return "unknown"
	}
}

var str2Type = map[string]Type{
	"scientific":      Scientific,
	"decimal":         Decimal,
	"engineering":     Engineering,
	"architectural":   Architectural,
	"fractional":      Fractional,
	"windows desktop": WindowsDesktop,
	"windowsdesktop":  WindowsDesktop,
}

func TypeFromString(str string) (Type, bool) {
	t, ok := str2Type[strings.TrimSpace(strings.ToLower(str))]
	return t, ok
}
