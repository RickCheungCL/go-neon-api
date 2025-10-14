package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rick/go-neon-api/internal/http/handlers"
)

func NewRouter(h *handlers.Handlers) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // or limit to specific origins later
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization", "X-User-Id", "X-User-Role"},
	}))
	api := r.Group("/api")
	{
		// ----- Cases (READ-ONLY) -----
		api.GET("/cases", h.ListCases)   // admin: all, user: own (based on headers/middleware)
		api.GET("/cases/:id", h.GetCase) // details for a single case

		// ----- On-Site Visit (READ + MUTATIONS on subresources) -----
		api.GET("/cases/:id/onsite", h.GetOnSiteVisit)     // fetch visit + rooms tree (read)
		api.POST("/cases/:id/onsite", h.EnsureOnSiteVisit) // ensure visit exists for a case (optional but handy)

		// Rooms within an On-Site Visit
		api.POST("/onsite/:visitId/rooms", h.CreateRoom) // create a room for this visit
		api.PUT("/rooms/:roomId", h.UpdateRoom)          // update room fields
		api.DELETE("/rooms/:roomId", h.DeleteRoom)       // remove a room

		// Pickers
		api.GET("/products", h.ListProducts)
		api.GET("/lightfixturetypes", h.ListLightFixtureTypes)

		// Existing lighting in a room (CRUD)
		api.POST("/rooms/:roomId/existing", h.AddExistingProduct) // add existing fixture row
		api.PUT("/existing/:id", h.UpdateExistingProduct)         // update qty/flags/etc.
		api.DELETE("/existing/:id", h.DeleteExistingProduct)      // delete existing fixture row

		// Suggested lighting in a room (CRUD)
		api.POST("/rooms/:roomId/suggested", h.AddSuggestedProduct) // add suggested fixture row
		api.PUT("/suggested/:id", h.UpdateSuggestedProduct)         // update suggestion
		api.DELETE("/suggested/:id", h.DeleteSuggestedProduct)      // delete suggestion
	}

	return r
}
