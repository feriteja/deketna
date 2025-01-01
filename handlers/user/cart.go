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

// AddToCartHandler handles adding goods to the cart
// @Summary Add Product to Cart
// @Description Add a product with a specific quantity to the buyer's cart
// @Tags Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body AddToCartRequest true "Product ID and Quantity"
// @Success 200 {object} helper.SuccessResponse "Product added to cart successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid input data"
// @Failure 404 {object} helper.ErrorResponse "Product not found"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /cart [post]
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
		"cart_id":      cart.ID,
		"cart_item_id": cartItem.ID,
		"product":      product.Name,
		"quantity":     req.Quantity,
	})
}

// GetCarts retrieves all cart items for the logged-in user
// @Summary Get Cart Items
// @Description Retrieve all cart items for the logged-in user
// @Tags  Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 25)"
// @Success 200 {object} helper.PaginationResponse{data=[]CartItemResponse} "List of cart items"
// @Failure 400 {object} helper.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} helper.ErrorResponse "Failed to retrieve cart items"
// @Router /cart [get]
func GetCarts(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)

	buyerID := uint64(claims["userid"].(float64)) // Extract Buyer ID
	// Pagination parameters
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

	var cartItems []CartItemResponse
	var totalItems int64

	err = config.DB.Table("cart_items").
		Select(`
			cart_items.id,
			cart_items.product_id,
			cart_items.updated_at,
			cart_items.quantity,

			products.name AS product_name,
			products.price, products.image_url,
			(products.price * cart_items.quantity) AS total_price`).
		Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Joins("JOIN products ON products.id = cart_items.product_id").
		Where("carts.buyer_id = ?", buyerID).
		Order("cart_items.updated_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&cartItems).Error

	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to retrieve cart items"})
		return
	}

	config.DB.Model(&models.Cart{}).Where("buyer_id = ?", buyerID).Count(&totalItems)

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

	helper.SendPagination(c, http.StatusOK, "Cart items retrieved successfully", cartItems, pagination)
}

// DeleteCart deletes one or more items from the cart
// @Summary Delete Cart Items
// @Description Delete one or more items from the user's cart
// @Tags  Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body DeleteCartRequest true "Cart Item IDs to delete"
// @Success 200 {object} helper.SuccessResponse "Cart items deleted successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid input data"
// @Failure 500 {object} helper.ErrorResponse "Failed to delete cart items"
// @Router /cart [delete]
func DeleteCart(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)

	buyerID := uint64(claims["userid"].(float64)) // Extract Buyer ID

	var req DeleteCartRequest

	if err := c.ShouldBindJSON(&req); err != nil || len(req.CartItemIDs) == 0 {
		helper.SendError(c, http.StatusBadRequest, []string{"Invalid input data"})
		return
	}

	err := config.DB.Where("id IN ? AND cart_id IN (SELECT id FROM carts WHERE buyer_id = ?)", req.CartItemIDs, buyerID).
		Delete(&models.CartItem{}).Error

	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to delete cart items"})
		return
	}

	helper.SendSuccess(c, http.StatusOK, "Cart items deleted successfully", nil)
}

// UpdateCart updates the quantity of a cart item
// @Summary Update Cart Item
// @Description Update the quantity of a specific cart item
// @Tags  Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body UpdateCartRequest true "Cart item details"
// @Success 200 {object} helper.SuccessResponse "Cart item updated successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid input data"
// @Failure 500 {object} helper.ErrorResponse "Failed to update cart item"
// @Router /cart [put]
func UpdateCart(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)

	buyerID := uint64(claims["userid"].(float64)) // Extract Buyer ID
	var req UpdateCartRequest

	if err := c.ShouldBindJSON(&req); err != nil || req.Quantity < 1 {
		helper.SendError(c, http.StatusBadRequest, []string{"Invalid input data"})
		return
	}

	// Verify if the cart item belongs to the user
	var exists bool
	err := config.DB.
		Table("cart_items").
		Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.buyer_id = ?", req.CartItemID, buyerID).
		Select("1").
		Scan(&exists).Error

	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to update cart item"})
		return
	}
	if !exists {
		helper.SendError(c, http.StatusUnauthorized, []string{"Unauthorized access to cart item"})
		return
	}

	err = config.DB.Model(&models.CartItem{}).
		Where("id = ?", req.CartItemID).
		Update("quantity", req.Quantity).Error

	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to update cart item"})
		return
	}

	helper.SendSuccess(c, http.StatusOK, "Cart item updated successfully", nil)
}
