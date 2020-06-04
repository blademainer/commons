package collections

// Set
type set map[interface{}]interface{}

// NewSet 创建set
func NewSet() Set {
	return &set{}
}

func (set *set) Size() int {
	return len(*set)
}

// Add 添加元素
func (set *set) Add(elem interface{}) bool {
	_, ok := (*set)[elem]
	(*set)[elem] = true
	return ok
}

// Entries 返回结果集
func (set *set) Entries() []interface{} {
	rs := make([]interface{}, 0, len(*set))
	for k := range *set {
		rs = append(rs, k)
	}
	return rs
}
