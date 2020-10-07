package collections

type Collection interface {
	// Add 添加一个元素, 如果已经存在元素，则返回true
	Add(interface{}) bool

	// Entries 获取所有元素
	Entries() []interface{}

	// 元素个数
	Size() int
}

// Set 集合
type Set interface {
	Collection
}
