package easyctx

import (
	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/contextware/contentctx"
	"github.com/nstogner/contextware/entityctx"
	"github.com/nstogner/contextware/errorctx"
	"github.com/nstogner/contextware/httpctx"
	"github.com/nstogner/contextware/logctx"
	"github.com/nstogner/contextware/routerctx"
)

var MaxBytesSize = int64(1000000)

func Get(h httpctx.HandlerFunc) httprouter.Handle {
	return routerctx.Adapt(
		errorctx.Handle(
			logctx.Requests(
				logctx.Errors(
					contentctx.Request(
						contentctx.Response(
							h,
							contentctx.JsonAndXml,
						),
						contentctx.JsonAndXml,
					),
				),
			),
			false,
		),
	)
}

func Post(h httpctx.HandlerFunc, entityDef *entityctx.Definition) httprouter.Handle {
	return routerctx.Adapt(
		errorctx.Handle(
			logctx.Requests(
				logctx.Errors(
					contentctx.Request(
						contentctx.Response(
							entityctx.Unmarshal(
								entityctx.Validate(
									h,
									entityDef,
								),
								entityDef,
							),
							contentctx.JsonAndXml,
						),
						contentctx.JsonAndXml,
					),
				),
			),
			false,
		),
	)
}

func Put(h httpctx.HandlerFunc, entityDef *entityctx.Definition) httprouter.Handle {
	return Post(h, entityDef)
}

func Delete(h httpctx.HandlerFunc, entity interface{}) httprouter.Handle {
	return routerctx.Adapt(
		errorctx.Handle(
			logctx.Requests(
				logctx.Errors(
					contentctx.Response(
						h,
						contentctx.JsonAndXml,
					),
				),
			),
			false,
		),
	)
}
