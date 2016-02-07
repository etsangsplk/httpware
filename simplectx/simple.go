package simplectx

import (
	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/netmiddle/contentctx"
	"github.com/nstogner/netmiddle/errorctx"
	"github.com/nstogner/netmiddle/httpctx"
	"github.com/nstogner/netmiddle/routerctx"
)

var MaxBytesSize = int64(1000000)

func Get(h httpctx.HandlerFunc) httprouter.Handle {
	return routerctx.Adapt(
		errorctx.Handle(
			contentctx.Request(
				contentctx.Response(
					h,
					contentctx.JsonAndXml,
				),
				contentctx.JsonAndXml,
			),
		),
	)
}

func Post(h httpctx.HandlerFunc, entity interface{}) httprouter.Handle {
	return routerctx.Adapt(
		errorctx.Handle(
			contentctx.Request(
				contentctx.Response(
					contentctx.Unmarshal(
						h,
						entity,
						MaxBytesSize,
						nil,
					),
					contentctx.JsonAndXml,
				),
				contentctx.JsonAndXml,
			),
		),
	)
}

func Put(h httpctx.HandlerFunc, entity interface{}) httprouter.Handle {
	return Post(h, entity)
}

func Delete(h httpctx.HandlerFunc, entity interface{}) httprouter.Handle {
	return routerctx.Adapt(
		errorctx.Handle(
			contentctx.Response(
				h,
				contentctx.JsonAndXml,
			),
		),
	)
}
