package main

import "fmt"

//https://www.cnblogs.com/Dylansuns/p/6793032.html
func main() {
	list := []int{1, 345, 4, 34, 3, 6, 2, 6, 3, 56, 234, 6, 87, 89}
	var n = Node{
		data: 30,
	}
	for _, i := range list {
		fmt.Println("insert :", i)
		n.Add(i)
	}

	n.print()

	fmt.Println(n.Search(4).parent.data)
}

type Node struct {
	data                   int
	parent, lchild, rchild *Node
}

func (n *Node) print() {
	if n.lchild != nil {
		n.lchild.print()
	}
	fmt.Println("node: ", n.data)
	if n.rchild != nil {
		n.rchild.print()
	}
}

func (n *Node) Search(i int) *Node {
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
	p := n.parent

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
	t := n.Search(i)

	switch {
	case t == nil:
		return
	case t.lchild == nil && t.rchild == nil:
		switch {
		case n.data == t.data:
			n = nil
		case t.data == t.parent.lchild.data:
			t.parent.lchild = nil
			t.parent = nil
		case t.data == t.parent.rchild.data:
			t.parent.rchild = nil
			t.parent = nil
		}
	case t.lchild == nil && t.rchild != nil:
		switch {
		case n.data == t.data:
			n = t.rchild
		case t.data == t.parent.lchild.data:
			t.parent.lchild = t.rchild
			t.rchild.parent = t.parent
		case t.data == t.parent.rchild.data :
			t.parent.rchild = t.rchild
			t.rchild.parent = t.parent
		}
	case t.lchild !=nil && t.rchild!=nil:

	}

}
