package services

import (
	"net/http/httputil"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	HTTP_PROXY_TTL = 60
)

type HTTPProxyPool struct {
	sm     sync.Map
	timers sync.Map
	expire time.Duration
	r      *Resolver
}

func NewHTTPProxyPool(r *Resolver) *HTTPProxyPool {
	return &HTTPProxyPool{expire: time.Duration(HTTP_PROXY_TTL) * time.Second, r: r}
}

func (s *HTTPProxyPool) Get(src *Source, logger *logrus.Entry, invoke bool) (*httputil.ReverseProxy, error) {
	key := src.GetKey() + strconv.FormatBool(invoke)
	v, _ := s.sm.LoadOrStore(key, NewHTTPProxy(s.r, src, logger, invoke))
	t, tLoaded := s.timers.LoadOrStore(key, time.NewTimer(s.expire))
	timer := t.(*time.Timer)
	if !tLoaded {
		go func() {
			<-timer.C
			s.sm.Delete(key)
			s.timers.Delete(key)
		}()
	} else {
		timer.Reset(s.expire)
	}

	return v.(*HTTPProxy).Get()
}
