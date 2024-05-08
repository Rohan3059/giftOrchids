package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/controllers"
	"github.com/kravi0/BizGrowth-backend/middleware"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	// Use the provided gin.Engine instead of creating a new router
	//incomingRoutes.Use(cors.Default())
	router := gin.Default()
	router.Use(gin.Logger())
	config := cors.DefaultConfig()
    config.AllowAllOrigins = true
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Token","token"} // Add "Token" header
    incomingRoutes.Use(cors.New(config))
	incomingRoutes.GET("/search-suggestions", controllers.SuggestionsHandler())
	incomingRoutes.GET("/search-product", controllers.SearchProduct())
	incomingRoutes.GET("/getcategory", controllers.GetCategory())
	incomingRoutes.GET("/category",controllers.GetSingleCategory())
	incomingRoutes.PUT("/updatecategory", controllers.EditCategory())
	incomingRoutes.GET("/getproduct", controllers.GetProduct())
	incomingRoutes.GET("/product", controllers.SearchProductByQuery())
	incomingRoutes.POST("/update-user", controllers.UpdateUserDetails())
	incomingRoutes.POST("/post-requirement",controllers.CreateRequirementMessage())
	incomingRoutes.GET("/get-productReference",controllers.FetchProductsAndReferencesHandler())
	incomingRoutes.GET("/product-reference", controllers.GetProductReferenceHandler())
	
	incomingRoutes.GET("/all-attributesType",controllers.GetAllAttributes())
	incomingRoutes.GET("/get-attributeType",controllers.GetAttributeByID())
	
	incomingRoutes.GET("/product-reviews",controllers.GetProductReviews())
	incomingRoutes.GET("/approved-product-reviews",controllers.GetProductApprovedReviews())
	incomingRoutes.POST("/create-ticket",controllers.CreateTicket())
	incomingRoutes.GET("/get-feeds", controllers.GetAllFeedsHandler())

	incomingRoutes.GET("/ticket/:id",controllers.GetTicketById())

	
	incomingRoutes.POST("/add-admin",controllers.RegisterAdmin())
	
	//incomingRoutes.GET("/getseller", controllers.GetSeller())
	incomingRoutes.POST("/seller/reset-password", controllers.ResetPassword())
	
	incomingRoutes.POST("/validatesellerotp", controllers.LoginValidatePasswordOTP())
	
	incomingRoutes.POST("/sendOTP", controllers.SetOtpHandler())
	incomingRoutes.POST("/validate", controllers.ValidateOtpHandler())
	incomingRoutes.POST("/sellerOTPRegistration", controllers.SellerRegistrationSendOTP())
	incomingRoutes.POST("/validatesellerotpin", controllers.SellerRegistrationOtpVerification())
	incomingRoutes.POST("/seller/detailsUpdate", controllers.SellerRegistration())
	incomingRoutes.POST("/seller/registration",controllers.SellerEmailUpdate())
	incomingRoutes.POST("/seller/licenseDetailsUpdate",controllers.SellerLicenseUpdate())
	incomingRoutes.POST("/seller-login", controllers.SendLoginOTP())
	
	
	
	incomingRoutes.Use(middleware.UserAuthentication())
	incomingRoutes.POST("/product-enquiry", controllers.EnquiryHandler())
	incomingRoutes.GET("/get-enquiry", controllers.GetUserEnquiries())
	incomingRoutes.POST("/post-review",controllers.AddReviewHandler())
	incomingRoutes.GET("/load-user",controllers.LoadUser())

	
	incomingRoutes.Use(middleware.Authentication())
	incomingRoutes.POST("/admin/approveSeller", controllers.ApproveSeller())
	incomingRoutes.POST("/admin/addcategory", controllers.AddCategory())
	incomingRoutes.POST("/admin/category/approve/:id", controllers.ApproveCategory())
	incomingRoutes.PUT("/admin/updateProduct", controllers.UpdateProduct())
	incomingRoutes.POST("/admin/add-product", controllers.ProductViewerAdmin())
	incomingRoutes.GET("/admin/get-enquiry", controllers.GETEnquiryHandler())
	incomingRoutes.GET("/admin/enquiry/:id", controllers.GetAdminSingleEnquiry())
	incomingRoutes.GET("/admin/getseller", controllers.GetSeller())
	incomingRoutes.GET("/admin/approve-product", controllers.ApproveProduct())
	incomingRoutes.GET("/seller/info", controllers.LoadSeller())
	incomingRoutes.DELETE("/admin/delete-product", controllers.DeleteProduct())
	incomingRoutes.POST("/seller/update-product", controllers.AddProductReferenceHandler())
	incomingRoutes.POST("/seller/update-profilepicture",controllers.SellerUpdateProfilePictureHandler())
	incomingRoutes.POST("/admin/approve-review", controllers.ApproveReview())
	incomingRoutes.GET("/admin/all-reviews",controllers.GetReviews())
	incomingRoutes.POST("/admin/add-attributeType",controllers.AddAttributeType())
	incomingRoutes.GET("/admin/getTickets",controllers.GetTickets())
	
	incomingRoutes.POST("/admin/add-feed", controllers.PostFeedHandler())
	incomingRoutes.DELETE("/admin/delete-feed", controllers.DeleteFeed())


	//incomingRoutes.GET("/getcategory", controllers.GetCategory())
}
