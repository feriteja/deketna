package user

import (
	"deketna/config"
	"deketna/helper"
	"deketna/models"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// GetProducts retrieves a list of products
// @Summary Get Products
// @Description Retrieve a list of products available for users
// @Tags Products
// @Accept json
// @Produce json
// @Success 200 {object} helper.SuccessResponse{data=[]Product} "Products retrieved successfully"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /products [get]
func GetProducts(c *gin.Context) {
	var products []Product // Product is your GORM model struct

	// Fetch all products from the database
	if err := config.DB.Find(&products).Error; err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Failed to retrieve products", err.Error()})

		return
	}

	// Return the products as a JSON response
	helper.SendSuccess(c, http.StatusOK, "Products retrieved successfully", products)
}

// GetProductDetail retrieves the details of a specific product
// @Summary Get Product Detail
// @Description Retrieve detailed information of a specific product by its ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} helper.SuccessResponse{data=Product} "Product details"
// @Failure 400 {object} helper.ErrorResponse "Invalid input data"
// @Failure 404 {object} helper.ErrorResponse "Product not found"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /product/{id} [get]
func GetProductDetail(c *gin.Context) {
	// Parse product ID from the URL
	idParam := c.Param("id")
	productID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product Product // Product is your GORM model struct

	// Fetch the product by ID from the database
	if err := config.DB.First(&product, productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product details"})
		}
		return
	}

	// Return the product details as a JSON response
	c.JSON(http.StatusOK, product)
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
