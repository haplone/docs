package tree

//https://www.cnblogs.com/Dylansuns/p/6793032.html

type Node struct {
	data                   int
	parent, lchild, rchild *Node
}

func (n *Node) t() (r []int) {
	if n.lchild != nil {
		r = append(r, n.lchild.t()...)
	}
	r = append(r, n.data)
	if n.rchild != nil {
		r = append(r, n.rchild.t()...)
	}
	return
}

func (n *Node) Search(i int) *Node {
	//log.Printf("search node %d for %d\n", n.data, i)
	switch {
	case n.data == i:
		return n
	case n.data > i:
		return n.lchild.Search(i)
	case n.data < i:
		return n.rchild.Search(i)
	}
	return nil
}

func (n *Node) Add(i int) *Node {
	if n.data == i {
		return n
	}
	c := n
	var p *Node

	for {
		p = c
		if c.data == i {
			return c
		} else if c.data < i {
			c = c.rchild
		} else if c.data > i {
			c = c.lchild
		}

		if c == nil {
			break
		}
	}
	newNode := &Node{
		data:   i,
		parent: p,
	}
	if p.data > i {
		p.lchild = newNode
	} else if p.data < i {
		p.rchild = newNode
	}
	return newNode
}

func (n *Node) remove(i int) {
	//log.Printf("remove node %d\n", i)
	t := n.Search(i)
	//log.Printf("get node %d when remove \n", t.data)

	switch {
	case t == nil:
		return
	case t.lchild == nil && t.rchild == nil:
		switch {
		case n.data == t.data:
			n = nil
		case t.parent.lchild != nil && t.data == t.parent.lchild.data:
			t.parent.lchild = nil
			t.parent = nil
		case t.parent.rchild != nil && t.data == t.parent.rchild.data:
			t.parent.rchild = nil
			t.parent = nil
		}
	case t.lchild == nil && t.rchild != nil:
		switch {
		case n.data == t.data:
			n = t.rchild
		case t.parent.lchild != nil && t.data == t.parent.lchild.data:
			t.parent.lchild = t.rchild
			t.rchild.parent = t.parent
		case t.parent.rchild != nil && t.data == t.parent.rchild.data:
			t.parent.rchild = t.rchild
			t.rchild.parent = t.parent
		}
	case t.lchild != nil && t.rchild == nil:
		switch {
		case n.data == t.data:
			n = t.lchild
		case t == t.parent.lchild:
			t.parent.lchild = t.lchild
			t.lchild.parent = t.parent
		case t == t.parent.rchild:
			t.parent.rchild = t.lchild
			t.lchild.parent = t.parent
		}
	case t.lchild != nil && t.rchild != nil:
		var leftMaxNode *Node = t.lchild
		for ; leftMaxNode.rchild != nil; {
			leftMaxNode = leftMaxNode.rchild
		}
		leftMaxNode.rchild = t.rchild

		if t.parent != nil {
			leftMaxNode.parent = t.parent

			if t == t.parent.lchild {
				t.parent.lchild = leftMaxNode
			} else {
				t.parent.rchild = leftMaxNode
			}
		} else {
			n = leftMaxNode
		}

		if t.lchild != leftMaxNode {
			leftMaxNode.lchild = t.lchild
		}
		//leftMaxNode.rchild = t.rchild
		t.parent, t.lchild, t.rchild = nil, nil, nil
	}

}
