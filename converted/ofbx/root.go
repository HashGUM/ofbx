package ofbx

type Node struct {
	Object
	typ Type
}

func NewNode(scene *Scene, element *Element, typ Type) *Node {
	n := &Node{}
	n.Object = *NewObject(scene, element)
	n.is_node = true
	n.typ = typ
	return n
}

func (n *Node) Type() Type {
	return n.typ
}

func (n *Node) String() string {
	return n.typ.String() + ":" + n.Object.String()
}