package user

import (
	"deketna/config"
	"deketna/helper"
	"deketna/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GetProducts retrieves a paginated list of products with seller details
// @Summary Get Products
// @Description Retrieve a paginated list of products with seller details
// @Tags   Product
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 25)"
// @Success 200 {object} helper.PaginationResponse{data=[]ProductWithSeller} "List of products with seller details"
// @Failure 400 {object} helper.ErrorResponse "Invalid query parameters"
// @Router /products [get]
func GetProducts(c *gin.Context) {
	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "25")

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
		Offset(offset).
		Scan(&products)

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

// AddToCartHandler handles adding goods to the cart
// @Summary Add Product to Cart
// @Description Add a product with a specific quantity to the buyer's cart
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body AddToCartRequest true "Product ID and Quantity"
// @Success 200 {object} helper.SuccessResponse "Product added to cart successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid input data"
// @Failure 404 {object} helper.ErrorResponse "Product not found"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /cart/{id} [post]
func AddToCart(c *gin.Context) {
	// Retrieve claims from middleware (user ID)
	claims := c.MustGet("claims").(jwt.MapClaims)

	buyerID := uint64(claims["userid"].(float64)) // Extract Buyer ID

	// Parse request body
	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Invalid input data", err.Error()})
		return
	}

	// Check if product exists
	var product models.Product
	if err := config.DB.First(&product, req.ProductID).Error; err != nil {
		helper.SendError(c, http.StatusNotFound, []string{"Product not found"})
		return
	}

	// Check if the buyer already has a cart
	var cart models.Cart
	if err := config.DB.Where("buyer_id = ?", buyerID).First(&cart).Error; err != nil {
		// If no cart exists, create one
		cart = models.Cart{BuyerID: buyerID}
		config.DB.Create(&cart)
	}

	// Check if the product already exists in the cart
	var cartItem models.CartItem
	if err := config.DB.Where("cart_id = ? AND product_id = ?", cart.ID, req.ProductID).First(&cartItem).Error; err == nil {
		// If product exists, update the quantity
		cartItem.Quantity += req.Quantity
		config.DB.Save(&cartItem)
	} else {
		// Add new cart item
		cartItem = models.CartItem{
			CartID:    cart.ID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		config.DB.Create(&cartItem)
	}

	// Return success response
	// Return success response
	helper.SendSuccess(c, http.StatusOK, "Product added to cart successfully", gin.H{
		"cart_id":  cart.ID,
		"product":  product.Name,
		"quantity": req.Quantity,
	})
}
