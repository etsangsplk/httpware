package routerezy

import (
	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/ctxware/adp/routeradp"
	"github.com/nstogner/ctxware/lib/httpctx"
	"github.com/nstogner/ctxware/mdl/contentmdl"
	"github.com/nstogner/ctxware/mdl/entitymdl"
	"github.com/nstogner/ctxware/mdl/errormdl"
	"github.com/nstogner/ctxware/mdl/logmdl"
)

var MaxBytesSize = int64(1000000)

func Get(h httpctx.HandlerFunc) httprouter.Handle {
	return routeradp.Adapt(
		errormdl.Handle(
			logmdl.Requests(
				logmdl.Errors(
					contentmdl.Request(
						contentmdl.Response(
							h,
							contentmdl.JsonAndXml,
						),
						contentmdl.JsonAndXml,
					),
				),
			),
			false,
		),
	)
}

func Post(h httpctx.HandlerFunc, entityDef entitymdl.Definition) httprouter.Handle {
	return routeradp.Adapt(
		errormdl.Handle(
			logmdl.Requests(
				logmdl.Errors(
					contentmdl.Request(
						contentmdl.Response(
							entitymdl.Unmarshal(
								entitymdl.Validate(
									h,
									entityDef,
								),
								entityDef,
							),
							contentmdl.JsonAndXml,
						),
						contentmdl.JsonAndXml,
					),
				),
			),
			false,
		),
	)
}

func Put(h httpctx.HandlerFunc, entityDef entitymdl.Definition) httprouter.Handle {
	return Post(h, entityDef)
}

func Delete(h httpctx.HandlerFunc, entity interface{}) httprouter.Handle {
	return routeradp.Adapt(
		errormdl.Handle(
			logmdl.Requests(
				logmdl.Errors(
					contentmdl.Response(
						h,
						contentmdl.JsonAndXml,
					),
				),
			),
			false,
		),
	)
}
