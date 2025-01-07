package user

import (
	"deketna/config"
	"deketna/helper"
	"deketna/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetProducts retrieves a paginated list of products with seller details
// @Summary Get Products
// @Description Retrieve a paginated list of products with seller details
// @Tags   Product
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 25)"
// @Param search_product query string false "Search specific product by keyword (default: "botol")"
// @Success 200 {object} helper.PaginationResponse{data=[]ProductWithSeller} "List of products with seller details"
// @Failure 400 {object} helper.ErrorResponse "Invalid query parameters"
// @Router /products [get]
func GetProducts(c *gin.Context) {
	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "25")

	search_product := c.Query("search_product")
	var searchProduct *string
	if search_product != "" {
		searchProduct = &search_product
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 25
	}

	offset := (page - 1) * limit

	// Fetch products with seller details
	var products []ProductWithSeller
	var totalItems int64

	query := config.DB.Table("products").
		Select(`
			products.id, 
			products.name, 
			products.price, 
			products.stock, 
			products.image_url, 
			users.id AS seller_id, 
			CASE 
				WHEN users.id = 1 THEN 'Deketna'
				ELSE COALESCE(profiles.name, '')
			END AS seller_name `).
		Joins("JOIN users ON users.id = products.seller_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = users.id").
		Limit(limit).
		Offset(offset)

	if searchProduct != nil {
		query = query.Where("LOWER(products.name) ILIKE LOWER(?)", "%"+*searchProduct+"%")

	}

	query.Scan(&products)

	if query.Error != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to retrieve products"})
		return
	}

	// Get total count for pagination
	config.DB.Model(&models.Product{}).Count(&totalItems)

	// Build pagination metadata
	totalPages := (int(totalItems) + limit - 1) / limit
	pagination := helper.PaginationMetadata{
		Page:       page,
		Limit:      limit,
		TotalItems: int(totalItems),
		TotalPages: totalPages,
		IsNext:     page < totalPages,
		IsPrev:     page > 1,
	}

	// Send success response
	helper.SendPagination(c, http.StatusOK, "Products retrieved successfully", products, pagination)
}

// GetProductDetail retrieves details of a specific product with seller details
// @Summary Get Product Detail
// @Description Retrieve details of a specific product with seller information
// @Tags  Product
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} helper.SuccessResponse{data=ProductWithSeller} "Product details with seller information"
// @Failure 400 {object} helper.ErrorResponse "Invalid Product ID"
// @Failure 404 {object} helper.ErrorResponse "Product not found"
// @Router /product/{id} [get]
func GetProductDetail(c *gin.Context) {
	// Parse product ID from the path
	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Invalid product ID"})

		return
	}

	// Fetch product with seller details using LEFT JOIN
	var product ProductWithSeller

	err = config.DB.Table("products").
		Select(`
			products.id, 
			products.name, 
			products.price, 
			products.stock, 
			products.image_url, 
			users.id AS seller_id, 
			CASE 
				WHEN users.id = 1 THEN 'Deketna'
				ELSE COALESCE(profiles.name, '')
			END AS seller_name `).
		Joins("JOIN users ON users.id = products.seller_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = users.id").
		Where("products.id = ?", productID).
		Scan(&product).Error

	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to retrieve product details"})
		return
	}

	// If no product found
	if product.ID == 0 {
		helper.SendError(c, http.StatusNotFound, []string{"Product not found"})

		return
	}

	// Send success response
	helper.SendSuccess(c, http.StatusOK, "Product details retrieved successfully", product)
}
