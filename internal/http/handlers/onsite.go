package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucsky/cuid"
	"github.com/rick/go-neon-api/internal/db"
)

// -------------------- EnsureOnSiteVisit --------------------
// POST /api/cases/:id/onsite
// Creates an OnSiteVisit if missing and returns the visit header.
func (h *Handlers) EnsureOnSiteVisit(c *gin.Context) {
	caseID := c.Param("id")

	// Try find
	var head struct {
		ID        string    `json:"id"        gorm:"column:id"`
		CaseID    string    `json:"caseId"    gorm:"column:caseId"`
		CreatedAt time.Time `json:"createdAt" gorm:"column:createdAt"`
	}
	if err := db.DB.Raw(
		`SELECT "id","caseId","createdAt" FROM "OnSiteVisit" WHERE "caseId" = ? LIMIT 1`,
		caseID,
	).Scan(&head).Error; err == nil && head.ID != "" {
		c.JSON(http.StatusOK, head)
		return
	}

	// Create new with CUID
	newID := cuid.New()
	if err := db.DB.Raw(
		`INSERT INTO "OnSiteVisit" ("id","caseId","createdAt")
		 VALUES (?, ?, now())
		 RETURNING "id","caseId","createdAt"`,
		newID, caseID,
	).Scan(&head).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create visit failed"})
		return
	}

	c.JSON(http.StatusOK, head)
}

// -------------------- GetOnSiteVisit --------------------
// GET /api/cases/:id/onsite
// Returns visit header and rooms with existing/suggested products.
func (h *Handlers) GetOnSiteVisit(c *gin.Context) {
	caseID := c.Param("id")

	// Visit
	var visit struct {
		ID        string    `json:"id"        gorm:"column:id"`
		CaseID    string    `json:"caseId"    gorm:"column:caseId"`
		CreatedAt time.Time `json:"createdAt" gorm:"column:createdAt"`
	}
	if err := db.DB.Raw(
		`SELECT "id","caseId","createdAt" FROM "OnSiteVisit" WHERE "caseId" = ? LIMIT 1`,
		caseID,
	).Scan(&visit).Error; err != nil || visit.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// Rooms
	type Room struct {
		ID              string    `json:"id"              gorm:"column:id"`
		OnSiteVisitID   string    `json:"onSiteVisitId"   gorm:"column:onSiteVisitId"`
		Location        string    `json:"location"        gorm:"column:location"`
		LocationTagID   *string   `json:"locationTagId"   gorm:"column:locationTagId"`
		LightingIssue   string    `json:"lightingIssue"   gorm:"column:lightingIssue"`
		CustomerRequest string    `json:"customerRequest" gorm:"column:customerRequest"`
		MountingKitQty  string    `json:"mountingKitQty"  gorm:"column:mountingKitQty"`
		MotionSensorQty int       `json:"motionSensorQty" gorm:"column:motionSensorQty"`
		CreatedAt       time.Time `json:"createdAt"       gorm:"column:createdAt"`
		CeilingHeight   *int      `json:"ceilingHeight"   gorm:"column:ceilingHeight"`
		Existing        []any     `json:"existing"`  // fill below
		Suggested       []any     `json:"suggested"` // fill below
	}
	var rooms []Room
	if err := db.DB.Raw(
		`SELECT "id","onSiteVisitId","location","locationTagId","lightingIssue","customerRequest",
		        "mountingKitQty","motionSensorQty","createdAt","ceilingHeight"
		   FROM "OnSiteVisitRoom"
		  WHERE "onSiteVisitId" = ?
		  ORDER BY "createdAt" DESC`,
		visit.ID,
	).Scan(&rooms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load rooms"})
		return
	}

	// For each room, load existing + suggested
	type ExistingRow struct {
		ID            string  `json:"id"            gorm:"column:id"`
		ProductID     string  `json:"productId"     gorm:"column:productId"`
		ProductName   string  `json:"productName"   gorm:"column:productName"`
		ProductWatt   float64 `json:"wattage"       gorm:"column:wattage"`
		Quantity      int     `json:"quantity"      gorm:"column:quantity"`
		BypassBallast bool    `json:"bypassBallast" gorm:"column:bypassBallast"`
	}
	type SuggestedRow struct {
		ID        string   `json:"id"          gorm:"column:id"`
		ProductID string   `json:"productId"   gorm:"column:productId"` // stores LightFixtureType.id
		TypeName  string   `json:"typeName"    gorm:"column:typeName"`
		SKU       *string  `json:"sku"         gorm:"column:SKU"`
		ImageURL  *string  `json:"imageUrl"    gorm:"column:imageUrl"`
		Wattage   *float64 `json:"wattage"     gorm:"column:wattage"`
		Quantity  int      `json:"quantity"    gorm:"column:quantity"`
	}

	for i := range rooms {
		r := &rooms[i]

		var ex []ExistingRow
		if err := db.DB.Raw(
			`SELECT e."id", e."productId", p."name" AS "productName", p."wattage",
			        e."quantity", e."bypassBallast"
			   FROM "OnSiteExistingProduct" e
			   JOIN "Product" p ON p."id" = e."productId"
			  WHERE e."roomId" = ?
			  ORDER BY e."id"`,
			r.ID,
		).Scan(&ex).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load existing"})
			return
		}

		var sg []SuggestedRow
		if err := db.DB.Raw(
			`SELECT s."id",
					s."productId",                         -- this is LightFixtureType.id in your DB
					l."name"        AS "typeName",
					l."SKU",
					l."imageUrl",
					l."wattage",
					s."quantity"
			FROM "OnSiteSuggestedProduct" s
			JOIN "LightFixtureType" l ON l."id" = s."productId"
			WHERE s."roomId" = ?
			ORDER BY s."id"`,
			r.ID,
		).Scan(&sg).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load suggested"})
			return
		}

		// assign
		r.Existing = make([]any, len(ex))
		for i2 := range ex {
			r.Existing[i2] = ex[i2]
		}
		r.Suggested = make([]any, len(sg))
		for i2 := range sg {
			r.Suggested[i2] = sg[i2]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        visit.ID,
		"caseId":    visit.CaseID,
		"createdAt": visit.CreatedAt,
		"rooms":     rooms,
	})
}

// -------------------- Rooms --------------------

type CreateRoomReq struct {
	Location        string  `json:"location" binding:"required"`
	LocationTagId   *string `json:"locationTagId"`
	LightingIssue   string  `json:"lightingIssue"`
	CustomerRequest string  `json:"customerRequest"`
	MountingKitQty  string  `json:"mountingKitQty"`
	MotionSensorQty int     `json:"motionSensorQty"`
	CeilingHeight   *int    `json:"ceilingHeight"`
}

// POST /api/onsite/:visitId/rooms
func (h *Handlers) CreateRoom(c *gin.Context) {
	visitID := c.Param("visitId")
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate CUID
	newID := cuid.New()

	// Get full room data back
	var row struct {
		ID              string    `json:"id" gorm:"column:id"`
		OnSiteVisitID   string    `json:"onSiteVisitId" gorm:"column:onSiteVisitId"`
		Location        string    `json:"location" gorm:"column:location"`
		LocationTagID   *string   `json:"locationTagId" gorm:"column:locationTagId"`
		LightingIssue   string    `json:"lightingIssue" gorm:"column:lightingIssue"`
		CustomerRequest string    `json:"customerRequest" gorm:"column:customerRequest"`
		MountingKitQty  string    `json:"mountingKitQty" gorm:"column:mountingKitQty"`
		MotionSensorQty int       `json:"motionSensorQty" gorm:"column:motionSensorQty"`
		CreatedAt       time.Time `json:"createdAt" gorm:"column:createdAt"`
		CeilingHeight   *int      `json:"ceilingHeight" gorm:"column:ceilingHeight"`
	}

	err := db.DB.Raw(
		`INSERT INTO "OnSiteVisitRoom"
		 ("id","onSiteVisitId","location","locationTagId","lightingIssue","customerRequest",
		  "mountingKitQty","motionSensorQty","createdAt","ceilingHeight")
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, now(), ?)
		 RETURNING "id","onSiteVisitId","location","locationTagId","lightingIssue","customerRequest",
		           "mountingKitQty","motionSensorQty","createdAt","ceilingHeight"`,
		newID, visitID, req.Location, req.LocationTagId, req.LightingIssue, req.CustomerRequest,
		req.MountingKitQty, req.MotionSensorQty, req.CeilingHeight,
	).Scan(&row).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}

	// Return complete room object with empty product arrays
	c.JSON(http.StatusCreated, gin.H{
		"id":              row.ID,
		"onSiteVisitId":   row.OnSiteVisitID,
		"location":        row.Location,
		"locationTagId":   row.LocationTagID,
		"lightingIssue":   row.LightingIssue,
		"customerRequest": row.CustomerRequest,
		"mountingKitQty":  row.MountingKitQty,
		"motionSensorQty": row.MotionSensorQty,
		"createdAt":       row.CreatedAt,
		"ceilingHeight":   row.CeilingHeight,
		"existing":        []any{},
		"suggested":       []any{},
	})
}

// PUT /api/rooms/:roomId
func (h *Handlers) UpdateRoom(c *gin.Context) {
	roomID := c.Param("roomId")

	// Accept only known keys to avoid SQL injection
	allowed := map[string]bool{
		"location":        true,
		"locationTagId":   true,
		"lightingIssue":   true,
		"customerRequest": true,
		"mountingKitQty":  true,
		"motionSensorQty": true,
		"ceilingHeight":   true,
	}

	var patch map[string]any
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	set := []string{}
	args := []any{}
	for k, v := range patch {
		if allowed[k] {
			set = append(set, fmt.Sprintf(`"%s" = ?`, k))
			args = append(args, v)
		}
	}
	if len(set) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid fields"})
		return
	}
	args = append(args, roomID)

	sql := `UPDATE "OnSiteVisitRoom" SET ` + strings.Join(set, ", ") + ` WHERE "id" = ?`
	if err := db.DB.Exec(sql, args...).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.Status(http.StatusOK)
}

// DELETE /api/rooms/:roomId
// DELETE /api/rooms/:roomId
func (h *Handlers) DeleteRoom(c *gin.Context) {
	roomID := c.Param("roomId")

	// Start a transaction to ensure all deletes succeed or none do
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete related records first (in order of dependencies)

	// 1. Delete photos associated with this room
	if err := tx.Exec(`DELETE FROM "Photo" WHERE "roomId" = ?`, roomID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete photos: " + err.Error()})
		return
	}

	// 2. Delete suggested lights (replacement products)
	if err := tx.Exec(`DELETE FROM "SuggestedLight" WHERE "roomId" = ?`, roomID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete suggested lights: " + err.Error()})
		return
	}

	// 3. Delete existing lights (current products)
	if err := tx.Exec(`DELETE FROM "ExistingLight" WHERE "roomId" = ?`, roomID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete existing lights: " + err.Error()})
		return
	}

	// 4. Finally delete the room itself
	if err := tx.Exec(`DELETE FROM "OnSiteVisitRoom" WHERE "id" = ?`, roomID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete room: " + err.Error()})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// -------------------- Existing Products --------------------

type AddProductReq struct {
	ProductID     string `json:"productId" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required"`
	BypassBallast *bool  `json:"bypassBallast"` // existing only
}

// POST /api/rooms/:roomId/existing
func (h *Handlers) AddExistingProduct(c *gin.Context) {
	roomID := c.Param("roomId")

	var req AddProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bypass := false
	if req.BypassBallast != nil {
		bypass = *req.BypassBallast
	}

	var row struct {
		ID string `gorm:"column:id"`
	}
	if err := db.DB.Raw(
		`INSERT INTO "OnSiteExistingProduct" ("id","roomId","productId","quantity","bypassBallast")
		 VALUES (gen_random_uuid()::text, ?, ?, ?, ?)
		 RETURNING "id"`,
		roomID, req.ProductID, req.Quantity, bypass,
	).Scan(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": row.ID})
}

// PUT /api/existing/:id
func (h *Handlers) UpdateExistingProduct(c *gin.Context) {
	id := c.Param("id")
	allowed := map[string]bool{
		"productId":     true,
		"quantity":      true,
		"bypassBallast": true,
	}
	var patch map[string]any
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	set := []string{}
	args := []any{}
	for k, v := range patch {
		if allowed[k] {
			set = append(set, fmt.Sprintf(`"%s" = ?`, k))
			args = append(args, v)
		}
	}
	if len(set) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid fields"})
		return
	}
	args = append(args, id)
	sql := `UPDATE "OnSiteExistingProduct" SET ` + strings.Join(set, ", ") + ` WHERE "id" = ?`
	if err := db.DB.Exec(sql, args...).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.Status(http.StatusOK)
}

// DELETE /api/existing/:id
func (h *Handlers) DeleteExistingProduct(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Exec(`DELETE FROM "OnSiteExistingProduct" WHERE "id" = ?`, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.Status(http.StatusNoContent)
}

// -------------------- Suggested Products --------------------

// POST /api/rooms/:roomId/suggested
func (h *Handlers) AddSuggestedProduct(c *gin.Context) {
	roomID := c.Param("roomId")

	var req AddProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var row struct {
		ID string `gorm:"column:id"`
	}
	if err := db.DB.Raw(
		`INSERT INTO "OnSiteSuggestedProduct" ("id","roomId","productId","quantity")
		 VALUES (gen_random_uuid()::text, ?, ?, ?)
		 RETURNING "id"`,
		roomID, req.ProductID, req.Quantity,
	).Scan(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": row.ID})
}

// PUT /api/suggested/:id
func (h *Handlers) UpdateSuggestedProduct(c *gin.Context) {
	id := c.Param("id")
	allowed := map[string]bool{
		"productId": true,
		"quantity":  true,
	}
	var patch map[string]any
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	set := []string{}
	args := []any{}
	for k, v := range patch {
		if allowed[k] {
			set = append(set, fmt.Sprintf(`"%s" = ?`, k))
			args = append(args, v)
		}
	}
	if len(set) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid fields"})
		return
	}
	args = append(args, id)
	sql := `UPDATE "OnSiteSuggestedProduct" SET ` + strings.Join(set, ", ") + ` WHERE "id" = ?`
	if err := db.DB.Exec(sql, args...).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.Status(http.StatusOK)
}

// DELETE /api/suggested/:id
func (h *Handlers) DeleteSuggestedProduct(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Exec(`DELETE FROM "OnSiteSuggestedProduct" WHERE "id" = ?`, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.Status(http.StatusNoContent)
}
