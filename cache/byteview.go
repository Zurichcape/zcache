package cache

/**
 * @author: zurich
 * @Date: 2024/3/15 22:00
 */

// ByteView 包含一个不可变的buf
type ByteView struct {
	buf []byte
}

func (bv ByteView) Len() int {
	return len(bv.buf)
}

// ByteSlice 返回一个buf的拷贝
func (bv ByteView) ByteSlice() []byte {
	return cloneBytes(bv.buf)
}

// 将数据转换为字符串返回
func (bv ByteView) String() string {
	return string(bv.buf)
}

func cloneBytes(buf []byte) []byte {
	cb := make([]byte, len(buf))
	copy(cb, buf)
	return cb
}
