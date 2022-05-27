package DawnCache

// ByteView 保存不可变的字节缓存值
type ByteView struct {
	b []byte
}

// Len 实现 lru.Value 接口
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回一个 ByteView 数据的克隆切片，ByteView 只读，所以返回克隆切片防止外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String 返回 ByteView 对应的字符串
func (v ByteView) String() string {
	return string(v.b)
}

// cloneBytes 克隆数据
func cloneBytes(b []byte) []byte {
	clone := make([]byte, len(b))
	copy(clone, b)
	return clone
}
