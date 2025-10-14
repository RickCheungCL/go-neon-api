package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rick/go-neon-api/internal/db"
)

// ---- Cases ----

type CreateCaseReq struct {
	UserID               string `json:"userId" binding:"required"`
	CustomerName         string `json:"customerName" binding:"required"`
	ProjectDetails       string `json:"projectDetails"`
	SchoolName           string `json:"schoolName"`
	ContactPerson        string `json:"contactPerson"`
	EmailAddress         string `json:"emailAddress"`
	PhoneNumber          string `json:"phoneNumber"`
	SchoolAddress        string `json:"schoolAddress"`
	LightingPurpose      string `json:"lightingPurpose"`
	FacilitiesUsedIn     string `json:"facilitiesUsedIn"`
	InstallationService  string `json:"installationService"`
	OperationDaysPerYear int    `json:"operationDaysPerYear"`
	OperationHoursPerDay int    `json:"operationHoursPerDay"`
}

func (h *Handlers) ListCases(c *gin.Context) {
	userID := strings.TrimSpace(c.GetHeader("X-User-Id"))
	userRole := strings.ToUpper(strings.TrimSpace(c.GetHeader("X-User-Role")))
	if userID == "" || (userRole != "ADMIN" && userRole != "USER") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	isAdmin := userRole == "ADMIN"

	// Optional: support ?page=&limit=
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit < 1 || limit > 500 {
		limit = 100
	}
	offset := (page - 1) * limit

	type Row struct {
		ID             string    `json:"id"              gorm:"column:id"`
		CustomerName   string    `json:"customerName"    gorm:"column:customerName"`
		ProjectDetails string    `json:"projectDetails"  gorm:"column:projectDetails"`
		ContactPerson  string    `json:"contactPerson"   gorm:"column:contactPerson"`
		SchoolName     string    `json:"schoolName"      gorm:"column:schoolName"`
		EmailAddress   string    `json:"emailAddress"    gorm:"column:emailAddress"`
		Status         string    `json:"status"          gorm:"column:status"`
		CreatedAt      time.Time `json:"createdAt"       gorm:"column:createdAt"`
		UpdatedAt      time.Time `json:"updatedAt"       gorm:"column:updatedAt"`
		UserName       *string   `json:"-"               gorm:"column:user_name"`
		UserEmail      string    `json:"-"               gorm:"column:user_email"`
	}

	var rows []Row
	baseSQL := `
		SELECT
			c."id", c."customerName", c."projectDetails", c."contactPerson", c."schoolName",
			c."emailAddress", c."status", c."createdAt", c."updatedAt",
			u."name" AS user_name, u."email" AS user_email
		FROM "Case" c
		JOIN "User" u ON u."id" = c."userId"
	`
	where := ``
	args := []any{}

	if !isAdmin {
		where = `WHERE c."userId" = ?`
		args = append(args, userID)
	}

	orderLimit := ` ORDER BY c."createdAt" DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	if err := db.DB.Raw(baseSQL+where+orderLimit, args...).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching cases"})
		return
	}

	// Shape the response with nested user {name,email}, like your TS select
	type UserLite struct {
		Name  *string `json:"name"`
		Email string  `json:"email"`
	}
	type Item struct {
		ID             string    `json:"id"`
		CustomerName   string    `json:"customerName"`
		ProjectDetails string    `json:"projectDetails"`
		ContactPerson  string    `json:"contactPerson"`
		SchoolName     string    `json:"schoolName"`
		EmailAddress   string    `json:"emailAddress"`
		Status         string    `json:"status"`
		CreatedAt      time.Time `json:"createdAt"`
		UpdatedAt      time.Time `json:"updatedAt"`
		User           UserLite  `json:"user"`
	}

	out := make([]Item, 0, len(rows))
	for _, r := range rows {
		out = append(out, Item{
			ID:             r.ID,
			CustomerName:   r.CustomerName,
			ProjectDetails: r.ProjectDetails,
			ContactPerson:  r.ContactPerson,
			SchoolName:     r.SchoolName,
			EmailAddress:   r.EmailAddress,
			Status:         r.Status,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
			User: UserLite{
				Name:  r.UserName,
				Email: r.UserEmail,
			},
		})
	}

	c.JSON(http.StatusOK, out)
}

func (h *Handlers) GetCase(c *gin.Context) {
	id := c.Param("id")

	// --- 1) Load the case + user (Prisma table/column names) ---
	var head struct {
		ID             string    `json:"id"              gorm:"column:id"`
		CustomerName   string    `json:"customerName"    gorm:"column:customerName"`
		ProjectDetails string    `json:"projectDetails"  gorm:"column:projectDetails"`
		ContactPerson  string    `json:"contactPerson"   gorm:"column:contactPerson"`
		SchoolName     string    `json:"schoolName"      gorm:"column:schoolName"`
		EmailAddress   string    `json:"emailAddress"    gorm:"column:emailAddress"`
		PhoneNumber    string    `json:"phoneNumber"     gorm:"column:phoneNumber"`
		SchoolAddress  string    `json:"schoolAddress"   gorm:"column:schoolAddress"`
		Status         string    `json:"status"          gorm:"column:status"`
		CreatedAt      time.Time `json:"createdAt"       gorm:"column:createdAt"`
		UpdatedAt      time.Time `json:"updatedAt"       gorm:"column:updatedAt"`
		UserName       *string   `json:"-"               gorm:"column:user_name"`
		UserEmail      string    `json:"-"               gorm:"column:user_email"`
	}

	sqlCase := `
		SELECT
			c."id", c."customerName", c."projectDetails", c."contactPerson",
			c."schoolName", c."emailAddress", c."phoneNumber", c."schoolAddress",
			c."status", c."createdAt", c."updatedAt",
			u."name" AS user_name, u."email" AS user_email
		FROM "Case" c
		JOIN "User" u ON u."id" = c."userId"
		WHERE c."id" = ?
		LIMIT 1
	`
	if err := db.DB.Raw(sqlCase, id).Scan(&head).Error; err != nil || head.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// --- 2) Load documents for the case ---
	type DocRow struct {
		ID              string    `json:"id"            gorm:"column:id"`
		URL             string    `json:"url"           gorm:"column:url"`
		FileName        string    `json:"fileName"      gorm:"column:fileName"`
		CustomName      *string   `json:"customName"    gorm:"column:customName"`
		UploadedViaLink bool      `json:"uploadedViaLink" gorm:"column:uploadedViaLink"`
		CreatedAt       time.Time `json:"createdAt"     gorm:"column:createdAt"`
	}
	var docs []DocRow
	sqlDocs := `
		SELECT "id","url","fileName","customName","uploadedViaLink","createdAt"
		FROM "Document"
		WHERE "caseId" = ?
		ORDER BY "createdAt" DESC
	`
	if err := db.DB.Raw(sqlDocs, id).Scan(&docs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load documents"})
		return
	}

	// --- 3) Load photos for the case ---
	type PhotoRow struct {
		ID              string    `json:"id"            gorm:"column:id"`
		URL             string    `json:"url"           gorm:"column:url"`
		Comment         *string   `json:"comment"       gorm:"column:comment"`
		CustomName      *string   `json:"customName"    gorm:"column:customName"`
		UploadedViaLink bool      `json:"uploadedViaLink" gorm:"column:uploadedViaLink"`
		CreatedAt       time.Time `json:"createdAt"     gorm:"column:createdAt"`
	}
	var photos []PhotoRow
	sqlPhotos := `
		SELECT "id","url","comment","customName","uploadedViaLink","createdAt"
		FROM "Photo"
		WHERE "caseId" = ?
		ORDER BY "createdAt" DESC
	`
	if err := db.DB.Raw(sqlPhotos, id).Scan(&photos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load photos"})
		return
	}

	// --- 4) Assemble response (similar to your Next.js select) ---
	type UserLite struct {
		Name  *string `json:"name"`
		Email string  `json:"email"`
	}
	resp := gin.H{
		"id":             head.ID,
		"customerName":   head.CustomerName,
		"projectDetails": head.ProjectDetails,
		"contactPerson":  head.ContactPerson,
		"schoolName":     head.SchoolName,
		"emailAddress":   head.EmailAddress,
		"phoneNumber":    head.PhoneNumber,
		"schoolAddress":  head.SchoolAddress,
		"status":         head.Status,
		"createdAt":      head.CreatedAt,
		"updatedAt":      head.UpdatedAt,
		"user": UserLite{
			Name:  head.UserName,
			Email: head.UserEmail,
		},
		"documents": docs,
		"photos":    photos,
	}

	c.JSON(http.StatusOK, resp)
}
