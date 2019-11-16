package djson

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Node struct {
	IfaceP *interface{}
}

const (
	TypeUnknown = iota
	TypeNull    // for json null
	TypeBool
	TypeFloat
	TypeString
	TypeArray
	TypeObject
)

func String2Node(s string) (Node, error) {
	return Bytes2Node([]byte(s))
}

func Bytes2Node(b []byte) (Node, error) {
	var iface interface{}
	err := json.Unmarshal(b, &iface)
	if err != nil {
		return Node{}, err
	}
	return Node{IfaceP: &iface}, nil
}

func Node2Bytes(node Node) ([]byte, error) {
	return json.Marshal(*(node.IfaceP))
}

func (node Node) Type() int {
	switch (*(node.IfaceP)).(type) {
	case string:
		return TypeString
	case bool:
		return TypeBool
	case float64:
		return TypeFloat
	case []interface{}:
		return TypeArray
	case map[string]interface{}:
		return TypeObject
	}
	return 0
}

func (node Node) Value() interface{} {
	return *node.IfaceP
}

func (node Node) String() string {
	if node.Type() == TypeString {
		return node.Value().(string)
	}
	return ""
}

func (node Node) Float() float64 {
	if node.Type() == TypeFloat {
		return node.Value().(float64)
	}
	return 0
}

func (node Node) Bool() bool {
	if node.Type() == TypeBool {
		return node.Value().(bool)
	}
	return false
}

func (node Node) Array() []interface{} {
	if node.Type() == TypeArray {
		return node.Value().([]interface{})
	}
	return nil
}

func (node Node) Object() map[string]interface{} {
	if node.Type() == TypeObject {
		return node.Value().(map[string]interface{})
	}
	return nil
}

func (proot Node) Get(path string) (Node, error) {
	segs := strings.Split(path, ".")
	res := proot
	var err error
	for _, seg := range segs {
		res, err = res.GetChild(seg)
		if err != nil {
			return Node{}, err
		}
	}
	return res, nil
}

func (proot Node) GetChild(path string) (Node, error) {
	typ := proot.Type()
	root := *(proot.IfaceP)
	if typ != TypeArray && typ != TypeObject {
		return Node{}, fmt.Errorf("leaf node")
	}
	var node Node
	if typ == TypeArray {
		i, err := strconv.ParseInt(path, 10, 16)
		if err != nil {
			return Node{}, err
		} else {
			node.IfaceP = &root.([]interface{})[i]
		}
	} else if typ == TypeObject {
		if v, ok := root.(map[string]interface{})[path]; ok {
			node.IfaceP = &v
		} else {
			return Node{}, fmt.Errorf("key '%v' not found", path)
		}
	} else {
		return Node{}, fmt.Errorf("unknown case")
	}
	return node, nil
}

func (proot Node) Set(path string, value interface{}) error {
	i := strings.LastIndex(path, ".")
	childPath := path[i+1:]
	if i == -1 {
		return proot.SetChild(childPath, value)
	}
	parentPath := path[:i]
	parent, err := proot.Get(parentPath)
	if err != nil {
		return err
	}
	return parent.SetChild(childPath, value)
}

func (proot Node) SetChild(path string, value interface{}) error {
	typ := proot.Type()
	root := *(proot.IfaceP)
	if typ != TypeArray && typ != TypeObject {
		return fmt.Errorf("leaf node")
	}
	if typ == TypeArray {
		index, err := strconv.ParseInt(path, 10, 16)
		if err != nil {
			return err
		} else {
			root.([]interface{})[index] = value
			return nil
		}
	} else if typ == TypeObject {
		if _, ok := root.(map[string]interface{})[path]; ok {
			root.(map[string]interface{})[path] = value
			return nil
		} else {
			return fmt.Errorf("key '%v' not found", path)
		}
	}
	return fmt.Errorf("unknown case")
}

func (proot Node) Delete(path string) error {
	i := strings.LastIndex(path, ".")
	childPath := path[i+1:]
	if i == -1 {
		return proot.DeleteChild(childPath)
	}
	parentPath := path[:i]
	parent, err := proot.Get(parentPath)
	if err != nil {
		return err
	}
	return parent.DeleteChild(childPath)
}

func (proot Node) DeleteChild(path string) error {
	typ := proot.Type()
	root := *proot.IfaceP
	if typ != TypeArray && typ != TypeObject {
		return fmt.Errorf("leaf node")
	}
	if typ == TypeArray {
		index64, err := strconv.ParseUint(path, 10, 64)
		index := int(index64)
		if err != nil {
			return err
		} else {
			s := root.([]interface{})
			for i := index + 1; i < len(s); i++ {
				s[i-1] = s[i]
			}
			*proot.IfaceP = s[:len(s)-1]
			return nil
		}
	} else if typ == TypeObject {
		if _, ok := root.(map[string]interface{})[path]; ok {
			delete(root.(map[string]interface{}), path)
			return nil
		} else {
			return fmt.Errorf("key '%v' not found", path)
		}
	}
	return fmt.Errorf("unknown case")
}
