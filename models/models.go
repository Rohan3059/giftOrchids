package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	Product_ID       primitive.ObjectID   `bson:"_id"`
	Product_Name     string               `json:"product_name" validate:"required"`
	Price            string               `json:"price" validate:"required"`
	Image            []string             `json:"image" validate:"required"`
	Discription      string               `json:"discription" validate:"required"`
	Category         string               `json:"category" validate:"required"`
	SKU              string               `json:"sku" validate:"required"`
	Featured         bool                 `json:"featured"`
	Approved         bool                 `json:"approved"`
	SellerRegistered []string             `json:"sellerid"`
	Attributes       []AttributeValue     `json:"attributes"`
	PriceRange       []ProductPriceRange  `json:"pricerange"`
	Variant          []ProductVariant     `json:"variant"`
	Reviews          []primitive.ObjectID ` bson:"reviews" json:"reviews" `
	Created_at       time.Time            `json:"created_at" bson:"created_at" `
	Updated_at       time.Time            `json:"updated_at" bson:"updated_at" `
	IsArchived       bool                 `json:"isArchived" bson:"isArchived" `
	IsRejected       bool                 `json:"isRejected" bson:"isRejected"`
	RejectionNote    string               `json:"rejection_note" bson:"rejection_note" `
}

type AttributeType struct {
	ID             primitive.ObjectID `bson:"_id" json:"_id"`
	Attribute_Name string             `bson:"attribute_name" json:"attribute_name"`
	Attribute_Code string             `bson:"attribute_code" json:"attribute_code"`
	Options        []string           `bson:"options" json:"options"`
}
type AttributeValue struct {
	AttributeType primitive.ObjectID `bson:"attribute_type" json:"attribute_type"`
	Value         []string           `bson:"attribute_value" json:"attribute_value"`
}

type ProductVariant struct {
	Attribute []AttributeValue `bson:"attribute" json:"attribute"`
}

type ProductPriceRange struct {
	MinQuantity int    `bson:"minQuantity" json:"minQuantity"`
	MaxQuantity int    `bson:"maxQuantity" json:"maxQuantity"`
	Price       string `bson:"price" json:"price" `
}

type Reviews struct {
	Id             primitive.ObjectID `bson:"_id" json:"_id"`
	ProductId      primitive.ObjectID `bson:"product_id" json:"product_id"`
	ReviewsDetails ReviewsDetails     `bson:"reviews_details" json:"reviews_details"`
	UserId         primitive.ObjectID `bson:"user_id" json:"user_id"`
	Status         string             `bson:"status" json:"status"` // approved or pending
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
	Approved       bool               `bson:"approved" json:"approved"`
	Archived       bool               `bson:"archived" json:"archived"`
}

type ReviewsDetails struct {
	ReviewTitle  string `bson:"review_title" json:"review_title"`
	ReviewText   string `bson:"review_text" json:"review_text"`
	ReviewRating int    `bson:"review_rating" json:"review_rating"`
}

type ProductReference struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id"`
	ProductID   primitive.ObjectID `bson:"product_id" json:"product_id"`
	SellerID    primitive.ObjectID `bson:"seller_id" json:"seller_id"`
	Price       string             `bson:"price"`
	MinQuantity int                `bson:"minQuantity" `

	MaxQuantity int       `bson:"maxQuantity" json:"maxQuantity"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`

	Approved bool `bson:"approved" json:"approved"`
	Archived bool `bson:"archived" json:"archived"`
}

type Units struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	Unit_Name string             `bson:"unit_name" json:"unit_name"`
	Unit_Code string             `bson:"unit_code" json:"unit_code"`
	Value     []string           `bson:"value" json:"value"`
}

type Categories struct {
	Category_ID          primitive.ObjectID `bson:"_id"`
	Category             string             `json:"category" bson:"category"`
	Category_image       string             `json:"category_image" bson:"category_image"`
	Category_Description string             `json:"category_description" bson:"category_description"`
	Parent_Category      primitive.ObjectID `json:"parent_category" bson:"parent_category"`
	Approved             bool               `json:"isApproved" bson:"isApproved"`
	Archived             bool               `json:"isArchived" bson:"isArchived"`
	Created_at           time.Time          `json:"created_at" bson:"created_at"`
	Updated_at           time.Time          `json:"updated_at" bson:"updated_at"`
	IsFeatured           bool               `json:"isFeatured" bson:"isFeatured"`
}

type Seller struct {
	ID              primitive.ObjectID `bson:"_id"`
	Seller_ID       string             `bson:"seller_id"`
	Company_Name    string             `json:"company_name" bson:"company_name"`
	OwnerDetail     OwnerDetails       `json:"owner_details"`
	CompanyDetail   CompanyDetail      `json:"companydetail" validate:"required"`
	MobileNo        string             `json:"mobileno"`
	Email           string             `json:"email" validate:"required"`
	OTP             string             `json:"otp"`
	Address_Details Address            `json:"address" bson:"address"`
	Approved        bool               `json:"approved"`
	Password        string             `json:"password" validate:"required,min=6"`
	Token           string             `json:"token"`
	User_type       string             `json:"user_type" Vallidate:"required, eq=ADMIN|eq=SELLER"`
	Refresh_token   string             `json:"refresh_token"`
	Created_at      time.Time          `json:"created_at"`
	Updated_at      time.Time          `json:"updated_at"`
	IsArchived      bool               `json:"archived"`
}

type OwnerDetails struct {
	Name             string `json:"name" bson:"name" `
	Email            string `json:"email" bson:"email" `
	MobileNo         string `json:"mobileno" bson:"mobileno"`
	Gender           string `json:"gender" bson:"gender"`
	DateOfBirth      string `json:"dateofbirth" bson:"dateofbirth" `
	AadharNumber     string `json:"aadharNumber" bson:"aadharNumber"`
	HavePassport     bool   `json:"havePassport" bson:"havePassport"`
	PassportNo       string `json:"passportNo" bson:"passportNo"`
	AadharDocument   string `json:"aadharDoc" bson:"aadharDoc"`
	PassportDocument string `json:"passportDoc" bson:"passportDoc"`
	PAN              string `json:"pan" bson:"pan"`
	PanDocument      string `json:"panDoc" bson:"panDoc"`
}

type CompanyDetail struct {
	BusinessType         string            `json:"businesstype" validate:"required"`
	YearEstablished      string            `json:"yearestablished" validate:"required"`
	CompanyOrigin        string            `json:"companyorigin" validate:"required"`
	GSTIN                string            `json:"gstin" `
	CIN                  string            `json:"cin"`
	LLPIN                string            `json:"llpin"`
	BusinessEntity       string            `json:"businessentity"`
	NoOfEmployee         string            `json:"noofemployee"`
	HaveBusinessLicenses bool              `json:"havebusinesslicenses"`
	BusinessLicenses     []BusinessLicense `json:"businesslicenses"`
	HaveExportPermission bool              `json:"haveexportpermission"`
	ExportPermission     []BusinessLicense `json:"exportpermission"`
	PAN                  string            `json:"pan" validate:"required"`
	PermanentAddress     string            `json:"permanenetaddress" validate:"required"`
	ProfilePicture       string            `json:"profilepicture"`
	GSTINDoc             string            `json:"gstindoc"`
	CINDoc               string            `json:"cindoc"`
	LLPINDoc             string            `json:"llpindoc"`
	PANImage             string            `json:"pan_image" validate:"required"`
}

type BusinessLicense struct {
	LicenseName  string `json:"licensename"`
	LicenseValue string `json:"licensevalue"`
	IssuedDate   string `json:" issueddate"`
	LicenseFile  string `json:"licensefile"`
}

type SellerTmp struct {
	ID       primitive.ObjectID `bson:"_id"`
	MobileNo string             `json:"mobileno"`
	OTP      string             `json:"otp"`
}

type Address struct {
	Address_id primitive.ObjectID `bson:"_id"`
	House      *string            `json:"house" bson:"house"`
	Street     *string            `json:"street" bson:"street"`
	City       *string            `json:"city" bson:"city"`
	Pincode    *string            `json:"pincode" bson:"pincode"`
}

type USer struct {
	User_id       primitive.ObjectID `bson:"_id"`
	UserName      string             `json:"user_name" bson:"user_name" `
	MobileNo      string             `json:"mobileno" bson:"mobileno" `
	Email         string             `json:"email" bson:"email"`
	OTP           string             `json:"otp" bson:"otp" `
	Token         string             `json:"token" bson:"token" `
	Refresh_token string             `json:"refresh_token" bson:"refresh_token"`
	Created_at    time.Time          `json:"created_at" bson:"created_at" `
	Updated_at    time.Time          `json:"updated_at" bson:"updated_at"`
	User_Address  string             `json:"address" bson:"address"`
	IsArchived    bool               `json:"archived" bson:"archived"`
}

type Enquire struct {
	Enquire_id   primitive.ObjectID `bson:"_id" `
	EnquireId    string             `json:"enquire_id" bson:"enquire_id"`
	User_id      string             `json:"user_id" bson:"user_id"`
	Product_id   string             `json:"product_id" bson:"product_id"`
	Quantity     int                `json:"quantity" bson:"quantity" `
	Resolved     bool               ` json:"resolved" bson:"resolved" `
	Status       string             `json:"status" bson:"status" `
	Enquiry_note string             `json:"enquire_note" bson:"enquire_note" `
	Enquire_date time.Time          `json:"enquire_date" bson:"enquire_date"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

type RequirementMessage struct {
	Requirement_id primitive.ObjectID `bson:"_id"`
	IsResolved     bool               `json:"resolved"`
	IsDeleted      bool               `json:"deleted"`
	Name           string             `json:"name"`
	Email          string             `json:"email"`
	MobileNo       string             `json:"mobileno"`
	Message        string             `json:"message"`
	Created_at     time.Time          `json:"created_at"`
}

type CustomerSupportTicket struct {
	Ticket_id       primitive.ObjectID `bson:"_id"`
	TicketID        string             `json:"ticketid" bson:"ticketid"`
	IsResolved      bool               `json:"resolved" bson:"resolved"`
	IsDeleted       bool               `json:"deleted" bson:"deleted"`
	Name            string             `json:"name" bson:"name"`
	Email           string             `json:"email" bson:"email"`
	MobileNo        string             `json:"mobileno" bson:"mobileno"`
	Subject         string             `json:"subject" bson:"subject"`
	Message         string             `json:"message" bson:"message"`
	Status          string             `json:"status" bson:"status" `
	Attachments     []string           `json:"attachments" bson:"attachments"`
	ChatMessage     primitive.ObjectID `json:"chatmessage" bson:"chatmessage"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
	AssignedSupport string             `json:"assignedsupport" bson:"assignedsupport"`
}

type SupportChatMessage struct {
	SupportChatId   primitive.ObjectID `bson:"_id"`
	SupportTicketId string             `bson:"ticketId" json:"ticketId"`
	Messages        []ChatMessage      `bson:"messages" json:"messages"`
}

type ChatMessage struct {
	Message string    `json:"message" bson:"message"`
	Sender  string    `json:"sender" bson:"sender"`
	SentAt  time.Time `json:"sent_at" bson:"sent_at" `
}
