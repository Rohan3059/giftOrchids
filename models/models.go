package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	Product_ID       primitive.ObjectID `bson:"_id"`
	Product_Name     string             `json:"product_name" validate:"required"`
	Price            string             `json:"price" validate:"required"`
	Image            []string           `json:"image" validate:"required"`
	Discription      string             `json:"discription" validate:"required"`
	Category         string             `json:"category" validate:"required"`
	Brand            string             `json:"brand" validate:"required"`
	Color            []string           `json:"color"`
	SKU              string             `json:"sku" validate:"required"`
	Featured         bool               `json:"featured"`
	Approved         bool               `json:"approved"`
	SellerRegistered []string           `json:"sellerid"`
	ProductReference []ProductReference `json:"productReference"`
}

type ProductReference struct {
	ID primitive.ObjectID `bson:"_id" json:"_id"`
    ProductID primitive.ObjectID `bson:"product_id" json:"product_id"`
	SellerID  primitive.ObjectID `bson:"seller_id" json:"seller_id"`
	Price     string             `bson:"price"`
	MinQuantity int              `bson:"minQuantity" `
	
	MaxQuantity int                `bson:"maxQuantity" json:"maxQuantity"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`

	Approved bool `bson:"approved" json:"approved"`
	Archived bool `bson:"archived" json:"archived"`

    
}

type Categories struct {
	Category_ID primitive.ObjectID `bson:"_id"`
	Category    string             `json:"category" bson:"category"`
}

type Seller struct {
	ID              primitive.ObjectID `bson:"_id"`
	Seller_ID       string             `bson:"seller_id"`
	Company_Name    string             `json:"Company_name"`
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
type SellerTmp struct {
	ID       primitive.ObjectID `bson:"_id"`
	MobileNo string             `json:"mobileno"`
	OTP      string             `json:"otp"`
}
type CompanyDetail struct {
	NameOfOwner      string `json:"nameofowner" validate:"required"`
	AadharNumber     string `json:"aadharnumber" validate:"required"`
	PAN              string `json:"pan" validate:"required"`
	PermanentAddress string `json:"permanenetaddress" validate:"required"`

	ProfilePicture   string `json:"profile_picture"`

	AadharImage string `json:"aadhar_image" validate:"required"`
	PANImage    string `json:"pan_image" validate:"required"`
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
	UserName      string             `json:"user_name"`
	MobileNo      string             `json:"mobileno"`
	Email         string             `json:"email"`
	OTP           string             `json:"otp"`
	Token         string             `json:"token"`
	Refresh_token string             `json:"refresh_token"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
	User_Address  string             `json:"addres"`
	IsArchived    bool               `json:"archived"`
	
}

type Enquire struct {
	Enquire_id primitive.ObjectID `bson:"_id"`
	User_id    string             `bson:"user_id"`
	Product_id string             `bson:"product_id"`
	Quantity   int             `bson:"quantity"`
	Resolved   bool               `bson:"resolved"`
	Enquiry_note string 			`bson:"enquire_note"`	
}

type RequirementMessage struct{
	Requirement_id primitive.ObjectID `bson:"_id"`
	IsResolved     bool                `json:"resolved"`
	IsDeleted      bool                `json:"deleted"`
	Name           string              `json:"name"`
	Email          string              `json:"email"`
	MobileNo       string              `json:"mobileno"`
	Message        string              `json:"message"`
	

}
