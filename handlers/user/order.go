package user

import (
	"deketna/config"
	"deketna/helper"
	"deketna/models"
	"errors"
	"fmt"
	"log"
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

		return nil // Commit transaction if no errors
	})
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{err.Error()})
		return
	}

	// Success Response
	helper.SendSuccess(c, http.StatusOK, "Order placed successfully", gin.H{
		"message": "Order placed successfully",
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
