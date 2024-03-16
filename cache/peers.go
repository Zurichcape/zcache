package cache

import "zcache/cache/pb"

/**
 * @author: zurich
 * @Date: 2024/3/15 23:45
 */

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(req *pb.Request, res *pb.Response) error
}
