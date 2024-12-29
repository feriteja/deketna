package admin

import (
	"deketna/config"
	"deketna/helper"
	"deketna/models"
	"deketna/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// @Summary Add a product
// @Description Admin adds a new product
// @Tags Admin Product
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Product Name"
// @Param price formData number true "Product Price"
// @Param stock formData integer true "Product Stock"
// @Param image formData file true "Product Image"
// @Success 201 {object} helper.SuccessResponse "Product"
// @Success 400 {object} helper.ErrorResponse "Validation Error"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 403 {object} helper.ErrorResponse "Access forbidden"
// @Router /admin/product [post]
func AddProduct(c *gin.Context) {
	// Parse form data
	var req AddProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Handle file upload
	file, err := c.FormFile("image")
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Failed to upload image"})
		return
	}

	// Retrieve admin ID from JWT token
	claims := c.MustGet("claims").(jwt.MapClaims)
	adminID, exists := claims["userid"].(float64)
	if !exists {
		helper.SendError(c, http.StatusInternalServerError, []string{"User ID not found in claims"})
		return
	}

	// Save file temporarily
	tempFilePath := filepath.Join("tmp", file.Filename)
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Failed to save temporary image"})
		return
	}

	fmt.Println("tempFilePath", tempFilePath)
	// Upload to Supabase
	imageURL, err := utils.UploadImageToSupabase(tempFilePath, file.Filename)
	fmt.Print("imageURL", imageURL)
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{fmt.Sprintf("Failed to upload image: %v", err)})
		return
	}

	// Create product record in DB
	product := models.Product{
		Name:     req.Name,
		Price:    req.Price,
		Stock:    req.Stock,
		SellerID: uint64(adminID),
		ImageURL: imageURL,
	}

	if err := config.DB.Create(&product).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to add product to database"})
		return
	}

	// Respond with success
	helper.SendSuccess(c, http.StatusCreated, "Product added successfully", product)
}

func DeleteProduct(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// GetProducts retrieves a paginated list of products with seller details
// @Summary Get Products
// @Description Retrieve a paginated list of products with seller details
// @Tags   Admin Product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 25)"
// @Param seller_id query int false "id of seller (default: 1)"
// @Param seller_name query string false "Name of seller (default: Deketna)"
// @Param product_name query string false "Name of product (default: botol)"
// @Success 200 {object} helper.PaginationResponse{data=[]GetProductResponse} "List of products with seller details"
// @Failure 400 {object} helper.ErrorResponse "Invalid query parameters"
// @Router /admin/products [get]
func GetProduct(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))

	sellerIDParam := c.Query("seller_id")
	var sellerID *uint64
	if sellerIDParam != "" {
		id, err := strconv.ParseUint(sellerIDParam, 10, 64)
		if err == nil {
			sellerID = &id
		}
	}

	sellerName := c.Query("seller_name")
	var sellerNamePtr *string
	if sellerName != "" {
		sellerNamePtr = &sellerName
	}

	productName := c.Query("product_name")
	var productNamePtr *string
	if productName != "" {
		productNamePtr = &productName
	}

	products, totalItems, err := GetProductsPaginated(config.DB, page, limit, sellerID, sellerNamePtr, productNamePtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	totalPages := (int(totalItems) + limit - 1) / limit
	pagination := helper.PaginationMetadata{
		Page:       page,
		Limit:      limit,
		TotalItems: int(totalItems),
		TotalPages: totalPages,
		IsNext:     page < totalPages,
		IsPrev:     page > 1,
	}

	helper.SendPagination(c, http.StatusOK, "Products retrieved successfully", products, pagination)

}

func GetProductsPaginated(db *gorm.DB, page, limit int, sellerID *uint64, sellerName *string, productName *string) ([]GetProductResponse, int64, error) {
	var products []GetProductResponse
	var totalItems int64

	// Calculate offset
	offset := (page - 1) * limit

	query := db.Model(&models.Product{}).
		Preload("Seller").
		Preload("Category")

	if sellerID != nil {
		query = query.Where("products.seller_id = ?", *sellerID)
	}

	if sellerName != nil && *sellerName != "" {
		query = query.Joins("JOIN profiles ON profiles.user_id = products.seller_id").
			Where("LOWER(profiles.name) ILIKE LOWER(?)", "%"+*sellerName+"%")
	}

	if productName != nil && *productName != "" {
		query = query.Where("LOWER(name) ILIKE LOWER(?)", "%"+*productName+"%")
	}

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	userJSON, _ := json.MarshalIndent(products, "", "  ")
	fmt.Println(string(userJSON))

	// Map to DTO
	var response = make([]GetProductResponse, len(products))

	for i, product := range products {
		response[i] = GetProductResponse{
			ID:        product.ID,
			Name:      product.Name,
			Price:     product.Price,
			Stock:     product.Stock,
			SellerID:  product.SellerID,
			ImageURL:  product.ImageURL,
			CreatedAt: product.CreatedAt,
			UpdatedAt: product.UpdatedAt,
			Seller: Profile{
				ID:       product.Seller.ID,
				Name:     product.Seller.Name,     // Adjust if `Name` comes from profiles
				ImageURL: product.Seller.ImageURL, // Profiles image must be handled manually
			},
			Category: Category{
				ID:          product.Category.ID,
				Name:        product.Category.Name,
				Description: product.Category.Description,
			},
		}
	}

	return products, totalItems, nil
}
