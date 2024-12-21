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
