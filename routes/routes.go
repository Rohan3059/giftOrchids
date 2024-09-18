package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
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
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Token", "token"} // Add "Token" header
	incomingRoutes.Use(cors.New(config))
	incomingRoutes.Use(gzip.Gzip(gzip.DefaultCompression))
	incomingRoutes.GET("/search-suggestions", controllers.SuggestionsHandler())
	incomingRoutes.GET("/search-product", controllers.SearchProduct())
	incomingRoutes.GET("/getcategory", controllers.GetCategory())
	incomingRoutes.GET("/categories", controllers.GetCategoryTree())
	incomingRoutes.GET("/featured-category", controllers.GetFeaturedCategory())
	incomingRoutes.GET("/category", controllers.GetSingleCategory())
	incomingRoutes.PUT("/updatecategory", controllers.EditCategory())
	incomingRoutes.GET("/getproduct", controllers.GetProduct())
	incomingRoutes.GET("/product", controllers.SearchProductByQuery())
	incomingRoutes.POST("/update-user", controllers.UpdateUserDetails())
	incomingRoutes.POST("/post-requirement", controllers.CreateRequirementMessage())
	incomingRoutes.GET("/get-productReference", controllers.FetchProductsAndReferencesHandler())
	incomingRoutes.GET("/product-reference", controllers.GetProductReferenceHandler())
	incomingRoutes.GET("/featured-products", controllers.GetFeaturedProducts())
	incomingRoutes.GET("/blogs", controllers.GetPublishedBlogs())
	incomingRoutes.GET("/blog/get", controllers.GetBlogBySlug())
	incomingRoutes.GET("/all-attributesType", controllers.GetAllAttributes())
	incomingRoutes.GET("/get-attributeType/:id", controllers.GetAttributeByID())
	incomingRoutes.GET("/approved-product-reviews", controllers.GetProductApprovedReviews())
	incomingRoutes.POST("/create-ticket", controllers.CreateTicket())
	incomingRoutes.GET("/get-feeds", controllers.GetAllFeedsHandler())
	incomingRoutes.GET("/content/get-by-key/:contentKey", controllers.GetContentItemsByKey())

	incomingRoutes.POST("/add-admin", controllers.RegisterAdmin())

	incomingRoutes.POST("/seller/reset-password", controllers.ResetPassword())

	incomingRoutes.POST("/validatesellerotp", controllers.LoginValidatePasswordOTP())

	incomingRoutes.POST("/sendOTP", controllers.SetOtpHandler())
	incomingRoutes.POST("/validate", controllers.ValidateOtpHandler())
	incomingRoutes.POST("/sellerOTPRegistration", controllers.SellerRegistrationSendOTP())
	incomingRoutes.POST("/validatesellerotpin", controllers.SellerRegistrationOtpVerification())
	incomingRoutes.POST("/seller/detailsUpdate", controllers.SellerCommpanyDetailsUpdate())
	incomingRoutes.POST("/seller/update/owner-details", controllers.SellerOwnerDetailsUpdate())
	incomingRoutes.POST("/seller/registration", controllers.SellerEmailUpdate())
	incomingRoutes.POST("/seller/licenseDetailsUpdate", controllers.SellerLicenseUpdate())
	incomingRoutes.POST("/seller-login", controllers.SendLoginOTP())
	incomingRoutes.POST("/seller/verify-otp", controllers.SellerOtpVerfication())

	incomingRoutes.Use(middleware.UserAuthentication())
	incomingRoutes.POST("/product-enquiry", controllers.EnquiryHandler())
	incomingRoutes.GET("/get-enquiry", controllers.GetUserEnquiries())
	incomingRoutes.POST("/post-review", controllers.AddReviewHandler())
	incomingRoutes.GET("/load-user", controllers.LoadUser())
	incomingRoutes.PUT("/user/update-profile", controllers.UpdateUserProfile())

	incomingRoutes.Use(middleware.Authentication())
	incomingRoutes.GET("/seller/products", controllers.GetAllProductsForASellerHandler())
	incomingRoutes.POST("/seller/update/business-details", controllers.UpdateSellerBusinessDetails())
	incomingRoutes.POST("/seller/profile/update/owner-details", controllers.UpdateOwnerDetails())
	incomingRoutes.POST("/seller/update-product/:id", controllers.SellerUpdateProduct())
	incomingRoutes.POST("/seller/update-profilepicture", controllers.SellerUpdateProfilePictureHandler())
	incomingRoutes.GET("/seller/support-tickets", controllers.GetSellerSupportTicket())
	incomingRoutes.GET("/seller/info", controllers.LoadSeller())
	incomingRoutes.POST("/seller/ticket/chat/message/:id", controllers.AddSellerMessage())
	incomingRoutes.POST("/seller/confirm-password", controllers.SellerPasswordConfirmation())
	incomingRoutes.POST("/seller/update-password", controllers.UpdatePassword())
	incomingRoutes.POST("/admin/approveSeller", controllers.ApproveSeller())
	incomingRoutes.POST("/admin/addcategory", controllers.AddCategory())
	incomingRoutes.GET("/admin/categories", controllers.AdminGetCategoryHandler())
	incomingRoutes.POST("/admin/category/approve/:id", controllers.ApproveCategory())
	incomingRoutes.PUT("/admin/category/make-featured/:id", controllers.HandleCategoryFeatured())
	incomingRoutes.PUT("/admin/update-product/:id", controllers.UpdateProduct())
	incomingRoutes.POST("/admin/add-product", controllers.ProductViewerAdmin())
	incomingRoutes.POST("/admin/add-product/seller", controllers.AddProductByAdmin())
	incomingRoutes.GET("/admin/get-enquiry", controllers.GETEnquiryHandler())
	incomingRoutes.GET("/admin/enquiry/:id", controllers.GetAdminSingleEnquiry())
	incomingRoutes.POST("/admin/enquiry/update/status/:id", controllers.UpdateEnquireStatus())
	incomingRoutes.GET("/admin/getseller", controllers.GetSeller())
	incomingRoutes.GET("/admin/seller/products/:id", controllers.GetSellerProductForAdmin())
	incomingRoutes.GET("/admin/approve-product", controllers.ApproveProduct())
	incomingRoutes.PUT("/admin/reject-product/:id", controllers.RejectProduct())
	incomingRoutes.DELETE("/admin/delete-product", controllers.DeleteProduct())
	incomingRoutes.POST("/admin/approve-review/:id", controllers.ApproveReview())
	incomingRoutes.GET("/admin/all-reviews", controllers.GetReviews())
	incomingRoutes.GET("/admin/review/:id", controllers.GetReview())
	incomingRoutes.GET("/admin/product-reviews", controllers.GetProductReviews())
	incomingRoutes.POST("/admin/add-attributeType", controllers.AddAttributeType())
	incomingRoutes.PUT("/admin/update-attribute/:id", controllers.UpdateAttributeType())
	incomingRoutes.GET("/admin/getTickets", controllers.GetTickets())
	incomingRoutes.GET("/admin/tickets/count", controllers.GetTicketCounts())
	incomingRoutes.GET("/admin/ticket/:id", controllers.GetTicketById())
	incomingRoutes.POST("/admin/ticket/chat/message/:id", controllers.AddMessage())
	incomingRoutes.POST("/admin/ticket/update/status/:id", controllers.UpdateTicketStatus())
	incomingRoutes.GET("/admin/ticket/chat/messages/:id", controllers.GetChatMessagesHandler())

	incomingRoutes.POST("/admin/add-feed", controllers.PostFeedHandler())
	incomingRoutes.DELETE("/admin/delete-feed", controllers.DeleteFeed())
	incomingRoutes.POST("/admin/update-feed", controllers.UpdateFeed())
	incomingRoutes.GET("/admin/products", controllers.GetAllProducts())
	incomingRoutes.POST("/admin/updat/product/featured/:id", controllers.MakeProductFeatured())

	incomingRoutes.GET("/admin/dashboard/analytics", controllers.GetAnalytics())

	incomingRoutes.GET("/admin/all-users", controllers.GetUsersDetails_Admin())

	incomingRoutes.PUT("/admin/product/remove-image/:id", controllers.DeleteImageFromProduct())

	incomingRoutes.GET("/admin/requirement-messages", controllers.GetAllRequirementMessages())

	incomingRoutes.GET("/admin/requirement-message/:id", controllers.GetRequirementMessage())

	incomingRoutes.POST("/admin/add-reviews", controllers.AddReviewByAdmin())

	incomingRoutes.GET("/admin/load", controllers.LoadAdmin())

	incomingRoutes.GET("/admin/seller/doc/download", controllers.DownloadSellerDocs())

	incomingRoutes.GET("/admin/get-csv", controllers.GenerateCSVByCollection())

	incomingRoutes.POST("/admin/post-blog", controllers.CreateBlog())
	incomingRoutes.GET("/admin/blogs", controllers.GetBlogs())
	incomingRoutes.GET("/admin/blog/:id", controllers.GetBlogByID())
	incomingRoutes.POST("/admin/blog/publish/:id", controllers.PublishBlog())
	incomingRoutes.POST("/admin/blog/archive/:id", controllers.ArchiveBlog())
	incomingRoutes.DELETE("/admin/blog/delete/:id", controllers.DeleteBlog())
	incomingRoutes.POST("/admin/blog/update/:id", controllers.UpdateBlog())

	incomingRoutes.POST("/admin/content/create", controllers.CreateContentItem())
	incomingRoutes.POST("/admin/content/toggle-status/:id", controllers.ToggleContentItemStatus())
	incomingRoutes.POST("/admin/content/update-file-content/:contentKey", controllers.UpdateFileContentItemContent())
	incomingRoutes.POST("/admin/content/delete-file-content/:contentItemId/:index", controllers.DeleteImageFromContentItem())
	incomingRoutes.DELETE("/admin/content/delete/:id", controllers.DeleteContentItem())
	incomingRoutes.GET("/admin/contents", controllers.GetAllContentItems())
}
