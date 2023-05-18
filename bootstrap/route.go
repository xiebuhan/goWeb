package bootstrap

import (
	"github.com/gorilla/mux"
	"goWeb/pkg/route"
	"goWeb/routes"
)

// SetupRoute 路由初始化
func SetupRoute() *mux.Router  {
	router := mux.NewRouter()
	routes.RegisterWebRoutes(router)

	route.SetRoute(router)
	return  router
}