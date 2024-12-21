package user

import (
	"deketna/config"
	"deketna/helper"
	"deketna/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// PlaceOrder creates a new order for the authenticated buyer
// @Summary Place Order
// @Description Create a new order from the buyer's cart, validate stock, deduct quantities, and clear the cart
// @Tags User Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.SuccessResponse{data=object{order_id=uint64,total_amount=float64}} "Order placed successfully"
// @Failure 400 {object} helper.ErrorResponse "Bad Request: Cart is empty or insufficient stock"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /order [post]
func PlaceOrder(c *gin.Context) {
	// Step 1: Get Buyer Info
	claims := c.MustGet("claims").(jwt.MapClaims)

	buyerID := uint64(claims["userid"].(float64)) // Extract buyer ID

	var cart models.Cart
	if err := config.DB.Where("buyer_id = ?", buyerID).Find(&cart).Error; err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Cart is empty"})
		return
	}

	// Step 2: Get Cart Items
	var cartItems []models.CartItem
	if err := config.DB.Where("cart_id = ?", cart.ID).Find(&cartItems).Error; err != nil || len(cartItems) == 0 {
		helper.SendError(c, http.StatusBadRequest, []string{"Cart is empty"})
		return
	}

	// Step 3: Validate Stock
	var totalAmount float64
	var insufficientStock []string
	for _, item := range cartItems {
		var product models.Product
		if err := config.DB.First(&product, item.ProductID).Error; err != nil {
			insufficientStock = append(insufficientStock, "Product not found: "+strconv.FormatUint(item.ProductID, 10))
			continue
		}
		if product.Stock < item.Quantity {
			insufficientStock = append(insufficientStock, "Insufficient stock for product: "+product.Name)
			continue
		}
		totalAmount += float64(item.Quantity) * product.Price
	}

	if len(insufficientStock) > 0 {
		helper.SendError(c, http.StatusBadRequest, insufficientStock)
		return
	}

	// Step 4: Create Order
	order := models.Order{
		BuyerID:     buyerID,
		TotalAmount: totalAmount,
		Status:      "pending",
	}
	if err := config.DB.Create(&order).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to create order"})
		return
	}

	// Step 5: Create Order Items and Deduct Stock
	for _, item := range cartItems {
		var product models.Product
		if err := config.DB.First(&product, item.ProductID).Error; err != nil {
			insufficientStock = append(insufficientStock, "Product not found: "+strconv.FormatUint(item.ProductID, 10))
			continue
		}
		// Create Order Item
		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     float64(item.Quantity) * product.Price,
		}
		config.DB.Create(&orderItem)

		// Deduct stock
		config.DB.Model(&models.Product{}).Where("id = ?", item.ProductID).UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity))
	}

	// Step 6: Clear the Cart
	config.DB.Where("cart_id = ?", buyerID).Delete(&models.CartItem{})

	// Step 7: Return Success Response
	helper.SendSuccess(c, http.StatusOK, "Order placed successfully", gin.H{
		"order_id":     order.ID,
		"total_amount": totalAmount,
	})
}

// ViewOrders retrieves the orders for the authenticated buyer
// @Summary View Orders
// @Description Retrieve a list of orders placed by the authenticated buyer
// @Tags User Orders
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Security BearerAuth
// @Success 200 {object} helper.SuccessResponse{data=PaginatedOrdersResponse} "Paginated list of orders with details"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /orders [get]
func ViewOrders(c *gin.Context) {

	// Get buyer ID from JWT claims

	claims := c.MustGet("claims").(jwt.MapClaims)
	buyerID := uint64(claims["userid"].(float64)) // Extract buyer ID

	// Get pagination parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	// Convert page and limit to integers
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		limitInt = 10
	}
	offset := (pageInt - 1) * limitInt

	var totalItems int64
	if err := config.DB.Model(&models.Order{}).Where("buyer_id = ?", buyerID).Count(&totalItems).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to count orders"})
		return
	}

	// Fetch orders for the buyer
	var orders []models.Order
	if err := config.DB.Preload("Items.Product").Where("buyer_id = ?", buyerID).Find(&orders).Limit(limitInt).Offset(offset).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to retrieve orders"})
		return
	}

	// Format orders with detailed items
	formattedOrders := []gin.H{}
	for _, order := range orders {
		items := []gin.H{}
		for _, item := range order.Items {
			items = append(items, gin.H{
				"product_name": item.Product.Name,
				"quantity":     item.Quantity,
				"price":        item.Price,
			})
		}

		formattedOrders = append(formattedOrders, gin.H{
			"order_id":     order.ID,
			"total_amount": order.TotalAmount,
			"status":       order.Status,
			"items":        items,
		})
	}

	// Calculate pagination metadata
	isNext := (pageInt * limitInt) < int(totalItems)
	isPrev := pageInt > 1

	// Print pagination values

	// Fetch total number of orders for the buyer
	if err := config.DB.Model(&models.Order{}).Where("buyer_id = ?", buyerID).Count(&totalItems).Error; err != nil {
		log.Println("Error occurred while counting orders:", err)
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to count orders"})
		return
	}

	helper.SendSuccess(c, http.StatusOK, "Orders retrieved successfully", gin.H{
		"page":      pageInt,
		"limit":     limitInt,
		"totalItem": totalItems,
		"isNext":    isNext,
		"isPrev":    isPrev,
		"data":      formattedOrders,
	})
}
