package geecache

type Getter interface {
	Get(key string) ([]byte, error)
}

// 接口型函数
type GetterFunc func(key string) ([]byte, error)

func (gf GetterFunc) Get(key string) ([]byte, error) {
	return gf(key)
}




