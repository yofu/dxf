package dxf

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/format"
)

// ODAFileConverter is the path of the ODA File Converter executable
// (https://www.opendesign.com/guestfiles/oda_file_converter, free of
// charge, installed separately) used by ReadDWG, SaveAsDWG and ConvertDWG
// to read and write DWG files, as ezdxf's odafc add-on. If it is empty, the
// converter is searched in PATH and in the default install folders.
var ODAFileConverter = ""

// ErrNoODAFileConverter is returned if the ODA File Converter is not found.
var ErrNoODAFileConverter = errors.New("ODA File Converter not found (set dxf.ODAFileConverter)")

// odaVersions maps the versions of this package (and R12) to the version
// names of the converter.
var odaVersions = map[string]string{
	"AC1009":     "ACAD12",
	format.R2000: "ACAD2000",
	format.R2004: "ACAD2004",
	format.R2007: "ACAD2007",
	format.R2010: "ACAD2010",
	format.R2013: "ACAD2013",
	format.R2018: "ACAD2018",
}

// odaVersion returns the converter version name of v (AC10xx, R2000,
// ..., R2018, ACAD2000, ... or "" for the latest version).
func odaVersion(v string) (string, error) {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "" {
		return "ACAD2018", nil
	}
	if s, ok := odaVersions[v]; ok {
		return s, nil
	}
	for _, s := range odaVersions {
		if v == s || "R"+strings.TrimPrefix(s, "ACAD") == v {
			return s, nil
		}
	}
	return "", fmt.Errorf("unknown DWG version %q", v)
}

// odaArgs returns the command line arguments of the converter: convert the
// files matching filter in folder in into folder out.
func odaArgs(in, out, version, outType, filter string) []string {
	return []string{in, out, version, outType, "0", "1", filter}
}

func odaExecutable() (string, error) {
	if ODAFileConverter != "" {
		if _, err := os.Stat(ODAFileConverter); err != nil {
			return "", fmt.Errorf("%w: %v", ErrNoODAFileConverter, err)
		}
		return ODAFileConverter, nil
	}
	for _, name := range []string{"ODAFileConverter", "ODAFileConverter.exe"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	var patterns []string
	switch runtime.GOOS {
	case "windows":
		for _, pf := range []string{os.Getenv("ProgramFiles"), `C:\Program Files`} {
			if pf != "" {
				patterns = append(patterns, filepath.Join(pf, "ODA", "ODAFileConverter*", "ODAFileConverter.exe"))
			}
		}
	case "darwin":
		patterns = append(patterns, "/Applications/ODAFileConverter.app/Contents/MacOS/ODAFileConverter")
	default:
		patterns = append(patterns, "/usr/bin/ODAFileConverter", "/usr/local/bin/ODAFileConverter", "/opt/ODAFileConverter*/ODAFileConverter")
	}
	for _, pat := range patterns {
		ms, _ := filepath.Glob(pat)
		sort.Sort(sort.Reverse(sort.StringSlice(ms))) // the latest version first
		if len(ms) > 0 {
			return ms[0], nil
		}
	}
	return "", ErrNoODAFileConverter
}

// odaConvert converts file src to outType (DXF or DWG) of version in a new
// temporary folder and returns the path of the result; the caller removes
// the folder (filepath.Dir of the result).
func odaConvert(src, version, outType string) (string, error) {
	exe, err := odaExecutable()
	if err != nil {
		return "", err
	}
	ver, err := odaVersion(version)
	if err != nil {
		return "", err
	}
	in, err := os.MkdirTemp("", "dxf-oda-in")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(in)
	out, err := os.MkdirTemp("", "dxf-oda-out")
	if err != nil {
		return "", err
	}
	// the converter is given a folder: copy the file to an empty one with
	// an ASCII name
	ext := strings.ToLower(filepath.Ext(src))
	if err := copyFile(src, filepath.Join(in, "in"+ext)); err != nil {
		os.RemoveAll(out)
		return "", err
	}
	cmd := exec.Command(exe, odaArgs(in, out, ver, outType, "in"+ext)...)
	msg, err := cmd.CombinedOutput()
	result := filepath.Join(out, "in."+strings.ToLower(outType))
	if _, serr := os.Stat(result); serr != nil {
		os.RemoveAll(out)
		if err == nil {
			err = serr
		}
		// the converter writes problems to an .err file
		if b, rerr := os.ReadFile(filepath.Join(out, "in"+ext+".err")); rerr == nil {
			msg = append(msg, b...)
		}
		return "", fmt.Errorf("ODA File Converter: %v: %s", err, strings.TrimSpace(string(msg)))
	}
	return result, nil
}

func copyFile(src, dst string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()
	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, r); err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

// ReadDWG reads a DWG file with the ODA File Converter (converted to DXF
// R2018 first).
func ReadDWG(filename string) (*drawing.Drawing, error) {
	dxfFile, err := odaConvert(filename, format.R2018, "DXF")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(filepath.Dir(dxfFile))
	return FromFile(dxfFile)
}

// SaveAsDWG writes the drawing as DWG file of the version of the drawing
// with the ODA File Converter.
func SaveAsDWG(d *drawing.Drawing, filename string) error {
	tmp, err := os.MkdirTemp("", "dxf-oda-src")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	src := filepath.Join(tmp, "src.dxf")
	if err := d.SaveAs(src); err != nil {
		return err
	}
	dwg, err := odaConvert(src, d.Version(), "DWG")
	if err != nil {
		return err
	}
	defer os.RemoveAll(filepath.Dir(dwg))
	return copyFile(dwg, filename)
}

// ConvertDWG converts a DWG or DXF file src to dst (DWG or DXF, by the
// extension of dst) of version (R12, R2000, ..., R2018, AC10xx, or "" for
// the latest) with the ODA File Converter.
func ConvertDWG(src, dst, version string) error {
	outType := "DWG"
	if strings.EqualFold(filepath.Ext(dst), ".dxf") {
		outType = "DXF"
	}
	out, err := odaConvert(src, version, outType)
	if err != nil {
		return err
	}
	defer os.RemoveAll(filepath.Dir(out))
	return copyFile(out, dst)
}
