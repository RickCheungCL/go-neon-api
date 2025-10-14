package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

/*
Prisma → Go mapping notes
- String @id @default(cuid())  → string primary key; we auto-fill via BeforeCreate (cuid.New()).
- DateTime @default(now())     → time.Time with autoCreateTime/autoUpdateTime where appropriate.
- Float                         → float64
- Int                           → int
- uuid()                        → we model as text; if you prefer DB default, set type:uuid;default:gen_random_uuid().
*/

// ---------- Helpers ----------

type BaseStringID struct {
	ID string `gorm:"primaryKey;type:text" json:"id"`
}

func (b *BaseStringID) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = cuid.New()
	}
	return nil
}

// ---------- Enums ----------

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

// ---------- User ----------

type User struct {
	BaseStringID
	Email        string        `gorm:"uniqueIndex;not null" json:"email"`
	Name         *string       `json:"name,omitempty"`
	Password     *string       `json:"password,omitempty"`
	Role         Role          `gorm:"type:text;default:USER;not null" json:"role"`
	CreatedAt    time.Time     `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime" json:"updatedAt"`
	ActivityLogs []ActivityLog `gorm:"foreignKey:UserID;references:ID" json:"activityLogs,omitempty"`
	Cases        []Case        `gorm:"foreignKey:UserID;references:ID" json:"cases,omitempty"`
}

// ---------- Case ----------

type Case struct {
	BaseStringID
	UserID               string    `gorm:"index;not null" json:"userId"`
	CustomerName         string    `json:"customerName"`
	ProjectDetails       string    `json:"projectDetails"`
	UploadToken          string    `gorm:"type:text" json:"uploadToken"` // prisma default(uuid()); keep as text
	Status               string    `gorm:"default:New;not null" json:"status"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	SchoolName           string    `json:"schoolName"`
	ContactPerson        string    `json:"contactPerson"`
	EmailAddress         string    `json:"emailAddress"`
	PhoneNumber          string    `json:"phoneNumber"`
	SchoolAddress        string    `json:"schoolAddress"`
	Num2FtLinearHighBay  int       `gorm:"default:0" json:"num2FtLinearHighBay"`
	Num150WUFOHighBay    int       `gorm:"default:0" json:"num150WUFOHighBay"`
	Num240WUFOHighBay    int       `gorm:"default:0" json:"num240WUFOHighBay"`
	Num2x2LEDPanel       int       `gorm:"default:0" json:"num2x2LEDPanel"`
	Num2x4LEDPanel       int       `gorm:"default:0" json:"num2x4LEDPanel"`
	Num1x4LEDPanel       int       `gorm:"default:0" json:"num1x4LEDPanel"`
	Num4FtStripLight     int       `gorm:"default:0" json:"num4FtStripLight"`
	LightingPurpose      string    `json:"lightingPurpose"`
	FacilitiesUsedIn     string    `json:"facilitiesUsedIn"`
	InstallationService  string    `json:"installationService"`
	OperationDaysPerYear int       `gorm:"default:0" json:"operationDaysPerYear"`
	OperationHoursPerDay int       `gorm:"default:0" json:"operationHoursPerDay"`

	// Relations
	User               User                `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	ActivityLogs       []ActivityLog       `gorm:"foreignKey:CaseID;references:ID" json:"activityLogs,omitempty"`
	FixtureCounts      []CaseFixtureCount  `gorm:"foreignKey:CaseID;references:ID" json:"fixtureCounts,omitempty"`
	Documents          []Document          `gorm:"foreignKey:CaseID;references:ID" json:"documents,omitempty"`
	InstallationDetail *InstallationDetail `gorm:"foreignKey:CaseID;references:ID" json:"installationDetail,omitempty"`
	OnSiteVisit        *OnSiteVisit        `gorm:"foreignKey:CaseID;references:ID" json:"onSiteVisit,omitempty"`
	Photos             []Photo             `gorm:"foreignKey:CaseID;references:ID" json:"photos,omitempty"`
}

// ---------- ActivityLog ----------

type ActivityLog struct {
	BaseStringID
	CaseID    string    `gorm:"index;not null" json:"caseId"`
	Action    string    `json:"action"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UserID    string    `gorm:"index;not null" json:"userId"`
	Case      Case      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User      User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

// ---------- Photo / Document ----------

type Photo struct {
	BaseStringID
	URL             string    `json:"url"`
	CaseID          string    `gorm:"index;not null" json:"caseId"`
	UploadedViaLink bool      `gorm:"default:false" json:"uploadedViaLink"`
	Comment         *string   `json:"comment,omitempty"`
	CustomName      *string   `json:"customName,omitempty"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
	Case            Case      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type Document struct {
	BaseStringID
	URL             string    `json:"url"`
	FileName        string    `json:"fileName"`
	CustomName      *string   `json:"customName,omitempty"`
	CaseID          string    `gorm:"index;not null" json:"caseId"`
	UploadedViaLink bool      `gorm:"default:false" json:"uploadedViaLink"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
	Case            Case      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

// ---------- LightFixtureType / CaseFixtureCount ----------

type LightFixtureType struct {
	BaseStringID
	Name          string             `gorm:"uniqueIndex;not null" json:"name"`
	Description   *string            `json:"description,omitempty"`
	CreatedAt     time.Time          `gorm:"autoCreateTime" json:"createdAt"`
	SKU           *string            `json:"SKU,omitempty"`
	Wattage       *float64           `json:"wattage,omitempty"`
	ImageURL      *string            `json:"imageUrl,omitempty"`
	FixtureCounts []CaseFixtureCount `gorm:"foreignKey:FixtureTypeID;references:ID" json:"fixtureCounts,omitempty"`
}

type CaseFixtureCount struct {
	BaseStringID
	CaseID        string `gorm:"index;not null" json:"caseId"`
	FixtureTypeID string `gorm:"index;not null" json:"fixtureTypeId"`
	Count         int    `gorm:"default:0" json:"count"`

	Case        Case             `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	FixtureType LightFixtureType `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
}

// ---------- InstallationDetail / Tags / Pivot ----------

type InstallationDetail struct {
	BaseStringID
	CaseID        string                  `gorm:"uniqueIndex;not null" json:"caseId"`
	CeilingHeight *float64                `json:"ceilingHeight,omitempty"`
	Notes         *string                 `json:"notes,omitempty"`
	CreatedAt     time.Time               `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time               `gorm:"autoUpdateTime" json:"updatedAt"`
	Case          Case                    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Tags          []InstallationDetailTag `gorm:"foreignKey:InstallationDetailID;references:ID" json:"tags,omitempty"`
}

type InstallationTag struct {
	BaseStringID
	Name                   string                  `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt              time.Time               `gorm:"autoCreateTime" json:"createdAt"`
	InstallationDetailTags []InstallationDetailTag `gorm:"foreignKey:TagID;references:ID" json:"installationDetailTags,omitempty"`
}

type InstallationDetailTag struct {
	BaseStringID
	InstallationDetailID string             `gorm:"index;not null" json:"installationDetailId"`
	TagID                string             `gorm:"index;not null" json:"tagId"`
	InstallationDetail   InstallationDetail `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Tag                  InstallationTag    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
}

// ---------- OnSite Visit / Rooms / Products / Photos / Tags ----------

type OnSiteVisit struct {
	BaseStringID
	CaseID    string            `gorm:"uniqueIndex;not null" json:"caseId"`
	CreatedAt time.Time         `gorm:"autoCreateTime" json:"createdAt"`
	Case      Case              `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Rooms     []OnSiteVisitRoom `gorm:"foreignKey:OnSiteVisitID;references:ID" json:"rooms,omitempty"`
}

type Product struct {
	BaseStringID
	Name        string  `json:"name"`
	Wattage     float64 `json:"wattage"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`

	ExistingProducts []OnSiteExistingProduct `gorm:"foreignKey:ProductID;references:ID" json:"existingProducts,omitempty"`
}

type OnSiteVisitRoom struct {
	BaseStringID
	OnSiteVisitID   string    `gorm:"index;not null" json:"onSiteVisitId"`
	Location        string    `json:"location"`
	LocationTagID   *string   `gorm:"index" json:"locationTagId,omitempty"`
	LightingIssue   string    `json:"lightingIssue"`
	CustomerRequest string    `json:"customerRequest"`
	MountingKitQty  string    `json:"mountingKitQty"`
	MotionSensorQty int       `json:"motionSensorQty"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
	CeilingHeight   *int      `json:"ceilingHeight,omitempty"`

	ExistingLights  []OnSiteExistingProduct  `gorm:"foreignKey:RoomID;references:ID" json:"existingLights,omitempty"`
	SuggestedLights []OnSiteSuggestedProduct `gorm:"foreignKey:RoomID;references:ID" json:"suggestedLights,omitempty"`
	Photos          []OnSiteVisitPhoto       `gorm:"foreignKey:RoomID;references:ID" json:"photos,omitempty"`
	LocationTag     *OnSiteLocationTag       `json:"locationTag,omitempty"`
	OnSiteVisit     OnSiteVisit              `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type OnSiteExistingProduct struct {
	BaseStringID
	RoomID        string          `gorm:"index;not null" json:"roomId"`
	ProductID     string          `gorm:"index;not null" json:"productId"`
	Quantity      int             `json:"quantity"`
	BypassBallast bool            `gorm:"default:false" json:"bypassBallast"`
	Product       Product         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	Room          OnSiteVisitRoom `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type OnSiteSuggestedProduct struct {
	BaseStringID
	RoomID    string          `gorm:"index;not null" json:"roomId"`
	ProductID string          `gorm:"index;not null" json:"productId"`
	Quantity  int             `json:"quantity"`
	Room      OnSiteVisitRoom `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type OnSiteVisitPhoto struct {
	BaseStringID
	RoomID    string                     `gorm:"index;not null" json:"roomId"`
	URL       string                     `json:"url"`
	Comment   string                     `json:"comment"`
	CreatedAt time.Time                  `gorm:"autoCreateTime" json:"createdAt"`
	Room      OnSiteVisitRoom            `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Tags      []OnSiteVisitPhotoTagPivot `gorm:"foreignKey:PhotoID;references:ID" json:"tags,omitempty"`
}

type OnSitePhotoTag struct {
	BaseStringID
	Name         string                     `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt    time.Time                  `gorm:"autoCreateTime" json:"createdAt"`
	TaggedPhotos []OnSiteVisitPhotoTagPivot `gorm:"foreignKey:TagID;references:ID" json:"taggedPhotos,omitempty"`
}

type OnSiteVisitPhotoTagPivot struct {
	BaseStringID
	PhotoID string           `gorm:"index;not null" json:"photoId"`
	TagID   string           `gorm:"index;not null" json:"tagId"`
	Photo   OnSiteVisitPhoto `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Tag     OnSitePhotoTag   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
}

type OnSiteLocationTag struct {
	BaseStringID
	Name      string            `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt time.Time         `gorm:"autoCreateTime" json:"createdAt"`
	Rooms     []OnSiteVisitRoom `gorm:"foreignKey:LocationTagID;references:ID" json:"rooms,omitempty"`
}

// ---------- QuoteCounter / PaybackSetting ----------

type QuoteCounter struct {
	BaseStringID
	CaseID    string    `gorm:"uniqueIndex;not null" json:"caseId"`
	Count     int       `gorm:"default:1" json:"count"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type PaybackSetting struct {
	BaseStringID
	CaseID    string    `gorm:"uniqueIndex;not null" json:"caseId"`
	Value     float64   `json:"value"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
