package object

import (
	"fmt"
	"sort"

	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Dictionary represents DICTIONARY Object.
type Dictionary struct {
	handle int
	src    int            // handle in the file it was read from (kept when writing)
	owner  handle.Handler // nil: root dictionary (owner 0)
	item   map[string]handle.Handler
}

// IsObject is for Object interface.
func (d *Dictionary) IsObject() bool {
	return true
}

// NewDictionary creates a new Dictionary.
func NewDictionary() *Dictionary {
	ds := make(map[string]handle.Handler)
	d := &Dictionary{
		handle: 0,
		item:   ds,
	}
	return d
}

// Format writes data to formatter.
func (d *Dictionary) Format(f format.Formatter) {
	f.WriteString(0, "DICTIONARY")
	f.WriteHex(5, d.handle)
	if d.owner != nil {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteHex(330, d.owner.Handle())
		f.WriteString(102, "}")
		f.WriteHex(330, d.owner.Handle())
	} else {
		f.WriteString(330, "0")
	}
	f.WriteString(100, "AcDbDictionary")
	f.WriteInt(281, 1)
	keys := make([]string, len(d.item))
	{
		idx := 0
		for k := range d.item {
			keys[idx] = k
			idx++
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		f.WriteString(3, k)
		f.WriteHex(350, d.item[k].Handle())
	}
}

// String outputs data using default formatter.
func (d *Dictionary) String() string {
	f := format.NewASCII()
	return d.FormatString(f)
}

// FormatString outputs data using given formatter.
func (d *Dictionary) FormatString(f format.Formatter) string {
	d.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (d *Dictionary) Handle() int {
	return d.handle
}

// SetHandle sets a handle.
func (d *Dictionary) SetHandle(v *int) {
	d.handle = handle.Next(d.src, v)
	for _, val := range d.item {
		val.SetHandle(v)
	}
}

// AddItem adds new a new item to Dictionary.
func (d *Dictionary) AddItem(key string, value handle.Handler) error {
	if _, exist := d.item[key]; exist {
		return fmt.Errorf("key %s already exists", key)
	}
	d.item[key] = value
	return nil
}

// Item returns the item with key (nil if it does not exist).
func (d *Dictionary) Item(key string) handle.Handler {
	return d.item[key]
}

// SetOwner sets the owner dictionary.
func (d *Dictionary) SetOwner(h handle.Handler) {
	d.owner = h
}

// RemoveItem removes the item with key.
func (d *Dictionary) RemoveItem(key string) {
	delete(d.item, key)
}

// SourceHandle returns the handle in the file it was read from ("" if new).
func (d *Dictionary) SourceHandle() string {
	return handle.Format(d.src)
}

// SetSourceHandle sets the handle in the file it was read from;
// it is kept when the drawing is written.
func (d *Dictionary) SetSourceHandle(h string) {
	d.src = handle.Parse(h)
}
