# ezdxfcheck

Validation of this package with [ezdxf](https://ezdxf.mozman.at/).

1. Files written by this package are read and audited by ezdxf.
2. Files written by ezdxf are read by this package, written back
   (ASCII and binary), and compared with the originals by ezdxf.

## Usage

```sh
python -m venv venv
venv/bin/pip install ezdxf        # venv\Scripts\pip on Windows

# 1. files written by this package (ASCII, binary, Shift_JIS)
go run ./tools/ezdxfcheck generate out/gen
venv/bin/python tools/ezdxfcheck/audit.py out/gen

# 2. files written by ezdxf (R2000 ASCII/binary, Shift_JIS, R12)
venv/bin/python tools/ezdxfcheck/make_ezdxf.py out/ez
go run ./tools/ezdxfcheck roundtrip out/ez out/rt
venv/bin/python tools/ezdxfcheck/compare.py out/ez out/rt

# 3. transformations (mirror, non-uniform scaling) compared with ezdxf
go run ./tools/ezdxfcheck transform out/xf
venv/bin/python tools/ezdxfcheck/transform_check.py out/xf

# 4. paths and extents compared with ezdxf
go run ./tools/ezdxfcheck paths out/pa
venv/bin/python tools/ezdxfcheck/path_check.py out/pa
```

`transform_check.py` transforms the paths of the original entities by the
matrix of each transformation with ezdxf and compares them with the
entities transformed by this package (largest distance between them).

`audit.py` prints the result of `doc.audit()` (errors and fixes) and the
contents of each file. `compare.py` prints the audit result of the files
written back and the differences from the original files (entities and
their properties, layers, linetypes, styles, dimension styles, blocks,
groups, layouts and viewports).

Expected differences after a round trip:

- R12 files are written as AC1015 (R2000); layers get the lineweight
  DEFAULT (-3).
- The dimension style of `$DIMSTYLE` is added if the DIMSTYLE table does
  not have it.
