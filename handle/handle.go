// Handler interface
package handle

// Handler is interface for handle (code 5, 105, etc).
//     5: Entity handle; text string of up to 16 hexadecimal digits (fixed)
//     105: Object handle for DIMVAR symbol table entry
//     320-329: Arbitrary object handles; handle values that are taken "as is." They are not translated during INSERT and XREF operations
//     330-339: Soft-pointer handle; arbitrary soft pointers to other objects within same DXF file or drawing. Translated during INSERT and XREF operations
//     340-349: Hard-pointer handle; arbitrary hard pointers to other objects within same DXF file or drawing. Translated during INSERT and XREF operations
//     350-359: Soft-owner handle; arbitrary soft ownership links to other objects within same DXF file or drawing. Translated during INSERT and XREF operations
//     360-369: Hard-owner handle; arbitrary hard ownership links to other objects within same DXF file or drawing. Translated during INSERT and XREF operations
//     390-399: String representing handle value of the PlotStyleName object, basically a hard pointer, but has a different range to make backward compatibility easier to deal with.
//     Stored and moved around as an Object ID (a handle in DXF files) and a special type in AutoLISP. Custom non-entity objects may use the full range,
//     but entity classes only use 391-399 DXF group codes in their representation, for the same reason as the Lineweight range above.
type Handler interface {
	Handle() int
	SetHandle(*int)
}
