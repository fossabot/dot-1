package attributes

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Map attribute map of graph component attributes
type Map map[Key]fmt.Stringer

// Attributes graph component attributes data
type Attributes struct {
	attributes Map
}

// NewAttributes creates an empty attributes map
func NewAttributes() *Attributes {
	return &Attributes{
		attributes: make(Map),
	}
}

// GetAttribute returns a given attribute by its key
func (dotObjectData *Attributes) GetAttribute(key Key) fmt.Stringer {
	return dotObjectData.attributes[key]
}

// GetAttributes returns all current attributes
func (dotObjectData *Attributes) GetAttributes() Map {
	return dotObjectData.attributes
}

// Write transforms attributes into dot notation and writes on the given writer
func (dotObjectData *Attributes) Write(device io.Writer, mustBracket bool) {
	if len(dotObjectData.attributes) == 0 {
		return
	}

	if mustBracket {
		fmt.Fprint(device, "[")
	}
	first := true
	// first collect keys
	keys := []Key{}
	for k := range dotObjectData.attributes {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return strings.Compare(string(keys[i]), string(keys[j])) < 0
	})

	for _, k := range keys {
		if !first {
			if mustBracket {
				fmt.Fprint(device, ",")
			} else {
				fmt.Fprint(device, ";")
			}
		}
		switch attributeData := dotObjectData.attributes[k].(type) {
		case *HTML:
			fmt.Fprintf(device, "%s=<%s>", k, attributeData.value)
		case *Literal:
			fmt.Fprintf(device, "%s=%s", k, attributeData.value)
		default:
			fmt.Fprintf(device, "%s=%q", k, attributeData.String())
		}
		first = false
	}
	if mustBracket {
		fmt.Fprint(device, "]")
	} else {
		fmt.Fprint(device, ";")
	}
}

// SetAttribute defines the attribute value for the given key
func (dotObjectData *Attributes) SetAttribute(key Key, value fmt.Stringer) {
	dotObjectData.attributes[key] = value
}

// SetAttributes sets multiple attribute values
func (dotObjectData *Attributes) SetAttributes(attributeMap Map) {
	for k, v := range attributeMap {
		dotObjectData.attributes[k] = v
	}
}

// DeleteAttribute removes an attribute, if it exists
func (dotObjectData *Attributes) DeleteAttribute(key Key) {
	delete(dotObjectData.attributes, key)
}
