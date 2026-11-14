package format

// Versions of DXF ($ACADVER) that this package writes.
const (
	R2000 = "AC1015"
	R2004 = "AC1018"
	R2007 = "AC1021"
	R2010 = "AC1024"
	R2013 = "AC1027"
	R2018 = "AC1032"
)

// Versions returns the versions this package writes.
func Versions() []string {
	return []string{R2000, R2004, R2007, R2010, R2013, R2018}
}

// AtLeast reports whether the formatter writes version v or later, so that
// group codes introduced in v may be written. A formatter without a
// version (e.g. used by String methods) writes everything.
func AtLeast(f Formatter, v string) bool {
	vf, ok := f.(interface{ Version() string })
	if !ok || vf.Version() == "" {
		return true
	}
	return vf.Version() >= v
}

// Version returns the DXF version being written ("" if not set).
func (f *ASCII) Version() string { return f.version }

// SetVersion sets the DXF version being written.
func (f *ASCII) SetVersion(v string) { f.version = v }

// Version returns the DXF version being written ("" if not set).
func (f *Binary) Version() string { return f.version }

// SetVersion sets the DXF version being written.
func (f *Binary) SetVersion(v string) { f.version = v }
