package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rick/go-neon-api/internal/db"
)

// ---------- GET /api/products ----------
// For EXISTING lighting picker (joins use "Product")
func (h *Handlers) ListProducts(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	category := strings.TrimSpace(c.Query("category"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	type Row struct {
		ID          string  `json:"id"          gorm:"column:id"`
		Name        string  `json:"name"        gorm:"column:name"`
		Wattage     float64 `json:"wattage"     gorm:"column:wattage"`
		Category    *string `json:"category"    gorm:"column:category"`
		Description *string `json:"description" gorm:"column:description"`
	}

	where := []string{}
	args := []any{}
	if q != "" {
		where = append(where, `("name" ILIKE ? OR "description" ILIKE ?)`)
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	if category != "" {
		where = append(where, `"category" = ?`)
		args = append(args, category)
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	var rows []Row
	sql := `
		SELECT "id","name","wattage","category","description"
		FROM "Product"
		` + whereSQL + `
		ORDER BY "name" ASC
		LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	if err := db.DB.Raw(sql, args...).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// ---------- GET /api/lightfixturetypes ----------
// For SUGGESTED lighting picker (joins use "LightFixtureType")
func (h *Handlers) ListLightFixtureTypes(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	type Row struct {
		ID          string   `json:"id"          gorm:"column:id"`
		Name        string   `json:"name"        gorm:"column:name"`
		SKU         *string  `json:"sku"         gorm:"column:SKU"`
		Wattage     *float64 `json:"wattage"     gorm:"column:wattage"`
		ImageURL    *string  `json:"imageUrl"    gorm:"column:imageUrl"`
		Description *string  `json:"description" gorm:"column:description"`
		// If you later add categories/tags to fixture types, you can filter here too.
	}

	where := []string{}
	args := []any{}
	if q != "" {
		where = append(where, `("name" ILIKE ? OR "description" ILIKE ? OR "SKU" ILIKE ?)`)
		args = append(args, "%"+q+"%", "%"+q+"%", "%"+q+"%")
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	var rows []Row
	sql := `
		SELECT "id","name","SKU","wattage","imageUrl","description"
		FROM "LightFixtureType"
		` + whereSQL + `
		ORDER BY "name" ASC
		LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	if err := db.DB.Raw(sql, args...).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}
	c.JSON(http.StatusOK, rows)
}
