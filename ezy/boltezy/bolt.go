package boltezy

import (
	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/ctxware/adp/routeradp"
	"github.com/nstogner/ctxware/lib/httpctx"
	"github.com/nstogner/ctxware/mdl/boltmdl"
	"github.com/nstogner/ctxware/mdl/contentmdl"
	"github.com/nstogner/ctxware/mdl/entitymdl"
	"github.com/nstogner/ctxware/mdl/errormdl"
	"github.com/nstogner/ctxware/mdl/logmdl"
)

func Get(h httpctx.HandlerFunc, def boltmdl.Definition) httprouter.Handle {
	return routeradp.Adapt(
		errormdl.Handle(
			logmdl.Requests(
				logmdl.Errors(
					contentmdl.Request(
						contentmdl.Response(
							boltmdl.Get(
								h,
								def,
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

func Post(h httpctx.HandlerFunc, def boltmdl.Definition) httprouter.Handle {
	return routeradp.Adapt(
		errormdl.Handle(
			logmdl.Requests(
				logmdl.Errors(
					contentmdl.Request(
						contentmdl.Response(
							entitymdl.Unmarshal(
								entitymdl.Validate(
									boltmdl.Post(
										h,
										def,
									),
									def.EntityDef,
								),
								def.EntityDef,
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
