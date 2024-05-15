package utils

import "errors"

var (
	ErrorCantFindProduct    = errors.New("can't find the product")
	ErrorCantDecodeProducts = errors.New("can't find the product")
	ErrorUserIdIsInavlid    = errors.New("this user is not valid")
	ErrorCantAddProductCart = errors.New("cannot add this product to cart")
	ErrorCantRemoveItemCart = errors.New("cannot remove product from cart")
	ErrorCantGetItem        = errors.New("was unable to get the item from the cart")
	ErrorCantBuyCartItem    = errors.New("cannot update the purchase")
	ErrorCantUpdateUser     = errors.New("cannot update user cart")
)

//bson constant
const (
	Categories  = "category"
	Brand       = "brand"
	ProductName = "product_name"
	Mobileno    = "mobileno"
	BsonID      = "_id"
)

const (
	Initiated = "Initiated"

	InProgress = "In Progress"

	Closed = "Closed"
)

//user constant
const (
	Admin = "ADMIN"
)

const (
	Seller = "SELLER"
)
