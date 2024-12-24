package admin

import (
	"net/http"
	"strconv"

	"deketna/config"
	"deketna/helper"
	"deketna/models"

	"github.com/gin-gonic/gin"
)

// ViewOrders retrieves a paginated list of all orders for admin
// @Summary Admin View Orders
// @Description Retrieve a paginated list of all orders with buyer details
// @Tags Admin Orders
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param status query string false "Filter by order status (e.g., pending, completed, cancelled)"
// @Security BearerAuth
// @Success 200 {object} helper.SuccessResponse{data=PaginatedAdminOrdersResponse} "Paginated list of all orders with details"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /admin/orders [get]
func ViewOrders(c *gin.Context) {
	// Get pagination parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	status := c.Query("status") // Optional status filter

	// Convert pagination parameters to integers
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		limitInt = 10
	}
	offset := (pageInt - 1) * limitInt

	// Fetch total number of orders
	var totalItems int64
	query := config.DB.Model(&models.Order{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&totalItems).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to count orders"})
		return
	}

	// Fetch paginated orders with buyer and items details
	var orders []models.Order
	if err := query.Preload("Buyer").Preload("Items.Product").
		Offset(offset).
		Limit(limitInt).
		Find(&orders).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to retrieve orders"})
		return
	}

	// Format orders with detailed items and buyer information
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
			"buyer": gin.H{
				"id":    order.Buyer.ID,
				"email": order.Buyer.Email,
				"phone": order.Buyer.Phone,
			},
			"items": items,
		})
	}

	// Calculate pagination metadata
	isNext := (pageInt * limitInt) < int(totalItems)
	isPrev := pageInt > 1

	// Return the orders in a standardized format
	helper.SendSuccess(c, http.StatusOK, "Orders retrieved successfully", gin.H{
		"page":      pageInt,
		"limit":     limitInt,
		"totalItem": totalItems,
		"isNext":    isNext,
		"isPrev":    isPrev,
		"data":      formattedOrders,
	})
}

// UpdateOrderStatus updates the status of an order by admin
// @Summary Update Order Status
// @Description Admin can update the status of an order (accept, reject, ontheway, finish)
// @Tags Admin Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param status body object{status=string} true "New order status (accept, reject, ontheway, finish)"
// @Security BearerAuth
// @Success 200 {object} helper.SuccessResponse{data=object{order_id=uint64,status=string}} "Order status updated successfully"
// @Failure 400 {object} helper.ErrorResponse "Bad Request: Invalid status or order not found"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /admin/orders/{id}/status [put]
func UpdateOrderStatus(c *gin.Context) {
	// Parse order ID from path
	orderID := c.Param("id")

	// Parse new status from JSON body
	var req struct {
		Status string `json:"status" binding:"required,oneof=accept reject ontheway finish"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Invalid status provided"})
		return
	}

	// Fetch the order
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Order not found"})
		return
	}

	// Validate allowed status transitions (optional but recommended)
	validStatuses := map[string]bool{
		"accept":   true,
		"reject":   true,
		"ontheway": true,
		"finish":   true,
	}
	if !validStatuses[req.Status] {
		helper.SendError(c, http.StatusBadRequest, []string{"Invalid order status"})
		return
	}

	// Update order status
	order.Status = req.Status
	if err := config.DB.Save(&order).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to update order status"})
		return
	}

	// Return success response
	helper.SendSuccess(c, http.StatusOK, "Order status updated successfully", gin.H{
		"order_id": order.ID,
		"status":   order.Status,
	})
}
