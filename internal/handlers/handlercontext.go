//Package handlers for providing handlers
package handlers

import (
	"goblog/internal/cache"
	"goblog/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
)

// PongoTemplate
// Contains the precompiled versios of the pongo templates useful for enhancing performance of page
// rendering times by 5x
type PongoTemplate struct {
}

// HTTPHandlerContext provides context for passing global values to handlers
// such as http thread pools or database handlers
//
// SEE: https://drstearns.github.io/tutorials/gohandlerctx/
type HTTPHandlerContext struct {
	hConfig  *config.AppConfig
	dbClient *mongo.Client
	cache    *cache.Cache
}

//CTXHandlerContext constructs a new HandlerContext,
//ensuring that the dependencies are valid values
func CTXHandlerContext(config *config.AppConfig, dbclient *mongo.Client, cache *cache.Cache) *HTTPHandlerContext {

	return &HTTPHandlerContext{config, dbclient, cache}
}
