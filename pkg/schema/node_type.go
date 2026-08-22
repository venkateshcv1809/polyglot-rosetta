package schema

import "fmt"

type NodeType string

const (
	NodeTypeCategory    NodeType = "category"
	NodeTypeSubCategory NodeType = "sub-category"
	NodeTypeConcept     NodeType = "concept"
)

func (n NodeType) Validate() error {
	switch n {
	case NodeTypeCategory, NodeTypeSubCategory, NodeTypeConcept:
		return nil
	default:
		return fmt.Errorf("invalid node type %q", n)
	}
}
