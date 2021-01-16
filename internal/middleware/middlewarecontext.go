//Package middleware
package middleware

import (
	"github.com/rsvancara/goblog/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// HTTPHandlerContext provides context for passing global values to handlers
// such as http thread pools or database handlers
//
// SEE: https://drstearns.github.io/tutorials/gohandlerctx/
type MiddleWareContext struct {
	hConfig  *config.AppConfig
	dbClient *mongo.Client
}

//CTXHandlerContext constructs a new HandlerContext,
//ensuring that the dependencies are valid values
func CTXMiddlewareContext(config *config.AppConfig, dbclient *mongo.Client) *MiddleWareContext {

	return &MiddleWareContext{config, dbclient}
}
