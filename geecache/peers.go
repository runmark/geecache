package geecache

import "example.com/mark/geecache/geecachepb/geecachepb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

//type PeerGetter interface {
//	Get(group string, key string) ([]byte, error)
//}

type PeerGetter interface {
	Get(in *geecachepb.Request, out *geecachepb.Response) error
}