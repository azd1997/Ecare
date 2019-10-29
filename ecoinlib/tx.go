package ecoin

// UnixTimeStamp Unix时间戳
type UnixTimeStamp uint64

// Coin 数字货币
type Coin uint



// Serializer 序列化接口，本项目中block和tx实现了这个接口
type Serializer interface {
	Serialize() (result []byte, err error)
}

// Hasher 取哈希接口，本项目中block和tx实现了这个接口
type Hasher interface {
	Hash() (hash Hash, err error)
}















