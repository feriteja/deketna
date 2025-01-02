package user

import (
	"deketna/config"
	"deketna/helper"
	"deketna/models"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// PlaceOrder creates a new order for selected products
// @Summary Place Order
// @Description Create a new order with selected products, validate stock, deduct quantities
// @Tags User Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body []OrderItemRequest true "List of products and quantities"
// @Success 200 {object} helper.SuccessResponse{data=object{order_id=uint64,total_amount=float64}} "Order placed successfully"
// @Failure 400 {object} helper.ErrorResponse "Validation Error"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /order [post]
func PlaceOrder(c *gin.Context) {
	// Step 1: Get Buyer Info
	claims := c.MustGet("claims").(jwt.MapClaims)
	buyerID := uint64(claims["userid"].(float64)) // Extract buyer ID

	// Step 2: Parse and Validate Input
	var orderItems []OrderItemRequest
	if err := c.ShouldBindJSON(&orderItems); err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Invalid input: " + err.Error()})
		return
	}

	if len(orderItems) == 0 {
		helper.SendError(c, http.StatusBadRequest, []string{"No products selected for the order"})
		return
	}

	// Step 3: Begin Database Transaction
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var totalAmount float64
		var validOrderItems []models.OrderItem

		var insufficientStock []string

		for _, item := range orderItems {
			var product models.Product
			if err := tx.Set("gorm:query_option", "FOR UPDATE").
				First(&product, item.ProductID).Error; err != nil {
				insufficientStock = append(insufficientStock, fmt.Sprintf("product not found: %d", item.ProductID))
				continue
			}

			// Check stock availability
			if product.Stock < item.Quantity {
				insufficientStock = append(insufficientStock, fmt.Sprintf("insufficient stock for product: %s", product.Name))
				continue
			}

			totalAmount += float64(item.Quantity) * product.Price
			validOrderItems = append(validOrderItems, models.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     product.Price,
			})
		}

		// If there are any stock errors, return them
		if len(insufficientStock) > 0 {
			return errors.New(strings.Join(insufficientStock, "; "))
		}

		// Create Order
		order := models.Order{
			BuyerID:     buyerID,
			TotalAmount: totalAmount,
			Status:      "pending",
		}
		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("failed to create order: %v", err)
		}

		// Create Order Items and Deduct Stock
		for _, item := range validOrderItems {
			item.OrderID = order.ID
			if err := tx.Create(&item).Error; err != nil {
				return fmt.Errorf("failed to create order item: %v", err)
			}

			// Deduct stock
			if err := tx.Model(&models.Product{}).
				Where("id = ?", item.ProductID).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
				return fmt.Errorf("failed to deduct stock for product ID: %d", item.ProductID)
			}
		}

		var productIDs []uint64
		for _, item := range orderItems {
			productIDs = append(productIDs, item.ProductID)
		}

		if err := tx.Where("product_id IN ?", productIDs).
			Delete(&models.CartItem{}).Error; err != nil {
			return fmt.Errorf("failed to remove items from cart: %v", err)
		}

		return nil // Commit transaction if no errors
	})
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{err.Error()})
		return
	}

	// Success Response
	helper.SendSuccess(c, http.StatusOK, "Order placed successfully", gin.H{
		"message": "Order  placed successfully",
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
// @Success 200 {object} helper.PaginationResponse{data=[]OrderResponse} "List of products with seller details"
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

	// Step 1: Fetch Orders
	var orders []OrderResponse
	err = config.DB.Table("orders").
		Select(`
			orders.id AS order_id,
			orders.buyer_id,
			orders.total_amount,
			orders.status,
			orders.created_at,
			orders.updated_at`).
		Where("orders.buyer_id = ?", buyerID).
		Order("orders.created_at DESC").
		Limit(limitInt).
		Offset(offset).
		Scan(&orders).Error

	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to fetch orders"})
		return
	}

	// Step 2: Fetch Order Items for Each Order
	var orderIDs []uint64
	for _, order := range orders {
		orderIDs = append(orderIDs, order.OrderID)
	}

	var items []OrderItemResponse
	if len(orderIDs) > 0 {
		err = config.DB.Table("order_items").
			Select(`
				order_items.order_id,
				products.name AS product_name,
				order_items.quantity,
				order_items.price`).
			Joins("JOIN products ON products.id = order_items.product_id").
			Where("order_items.order_id IN ?", orderIDs).
			Scan(&items).Error

		if err != nil {
			helper.SendError(c, http.StatusInternalServerError, []string{"Failed to fetch order items"})
			return
		}
	}

	// Step 3: Map Order Items to Their Respective Orders
	orderItemMap := make(map[uint64][]OrderItemResponse)
	for _, item := range items {
		orderItemMap[item.OrderID] = append(orderItemMap[item.OrderID], item)
	}

	var finalOrders []struct {
		OrderResponse
		Items []OrderItemResponse `json:"order_items"`
	}
	for _, order := range orders {
		finalOrders = append(finalOrders, struct {
			OrderResponse
			Items []OrderItemResponse `json:"order_items"`
		}{
			OrderResponse: order,
			Items:         orderItemMap[order.OrderID],
		})
	}

	totalPages := (int(totalItems) + limitInt - 1) / limitInt
	pagination := helper.PaginationMetadata{
		Page:       pageInt,
		Limit:      limitInt,
		TotalItems: int(totalItems),
		TotalPages: totalPages,
		IsNext:     pageInt < totalPages,
		IsPrev:     pageInt > 1,
	}

	// Fetch total number of orders for the buyer

	helper.SendPagination(c, http.StatusOK, "Orders retrieved successfully", finalOrders, pagination)
}
