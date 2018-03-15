package main

import (
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/cache"
	"github.com/dougEfresh/passwd-pot/service"
)

var eventResolver service.EventResolver
var geoCache *cache.Cache = cache.NewCache()

func resolveEvent(event api.Event) (error) {
	var ids []int64
	rId, _ := geoCache.Get(event.RemoteAddr)
	oId, _ := geoCache.Get(event.OriginAddr)
	if rId > 0 && oId > 0 {
		if e := eventResolver.MarkRemoteEvent(event.ID, rId); e != nil {
			return e
		}
		if e := eventResolver.MarkOriginEvent(event.ID, oId) ; e != nil {
			return e
		}
	}
	ids, err := eventResolver.ResolveEvent(event)
	if err == nil {
		geoCache.Set(event.RemoteAddr, ids[0])
		geoCache.Set(event.OriginAddr, ids[1])
	}
	return err
}
