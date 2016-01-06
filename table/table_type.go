package table

type TableType int

const (
	VPORT TableType = iota
	LTYPE
	LAYER
	STYLE
	VIEW
	UCS
	APPID
	DIMSTYLE
	BLOCK_RECORD
)

func TableTypeString(t TableType) string {
	switch t {
	case VPORT:
		return "VPORT"
	case LTYPE:
		return "LTYPE"
	case LAYER:
		return "LAYER"
	case STYLE:
		return "STYLE"
	case VIEW:
		return "VIEW"
	case UCS:
		return "UCS"
	case APPID:
		return "APPID"
	case DIMSTYLE:
		return "DIMSTYLE"
	case BLOCK_RECORD:
		return "BLOCK_RECORD"
	default:
		return ""
	}
}

func TableTypeValue(t string) TableType {
	switch t {
	case "VPORT":
		return VPORT
	case "LTYPE":
		return LTYPE
	case "LAYER":
		return LAYER
	case "STYLE":
		return STYLE
	case "VIEW":
		return VIEW
	case "UCS":
		return UCS
	case "APPID":
		return APPID
	case "DIMSTYLE":
		return DIMSTYLE
	case "BLOCK_RECORD":
		return BLOCK_RECORD
	default:
		return -1
	}
}
