package admin

import (
	"net/http"
	"strconv"
	"time"

	"deketna/config"
	"deketna/helper"
	"deketna/models"

	"github.com/gin-gonic/gin"
)

// ViewOrders retrieves the orders for the authenticated buyer
// @Summary View Orders
// @Description Retrieve a list of orders placed by the authenticated buyer
// @Tags Admin Orders
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Security BearerAuth
// @Success 200 {object} helper.PaginationResponse{data=[]OrderDetailWithItemsResponse} "List of products with seller details"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /admin/orders [get]
func ViewOrders(c *gin.Context) {
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
	if err := config.DB.Model(&models.Order{}).Count(&totalItems).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to count orders"})
		return
	}

	type OrderResponseTemp struct {
		OrderID     uint64  `json:"order_id"`
		TotalAmount float64 `json:"total_amount"`
		Status      string  `json:"status"`
		CreatedAt   string  `json:"created_at"`
		UpdatedAt   string  `json:"updated_at"`
		BuyerID     uint64  `json:"buyer_id"`
		BuyerEmail  string  `json:"buyer_email"`
		BuyerName   string  `json:"buyer_name"`
		BuyerPhone  string  `json:"buyer_phone"`
	}

	// Step 1: Fetch Orders with Buyer Info
	var orders []OrderResponseTemp
	err = config.DB.Table("orders").
		Select(`
		orders.id AS order_id,
		orders.total_amount,
		orders.status,
		orders.created_at,
		orders.updated_at,
		profiles.name as buyer_name,
		users.id AS buyer_id,
		users.email AS buyer_email,
		users.phone AS buyer_phone`).
		Joins("JOIN profiles ON profiles.user_id = orders.buyer_id").
		Joins("JOIN users ON users.id = profiles.user_id").
		Order("orders.created_at DESC").
		Limit(limitInt).
		Offset(offset).
		Scan(&orders).Error

	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to fetch orders with buyer details"})
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
				order_items.price,
				order_items.quantity * order_items.price AS total_price,
				products.image_url`).
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

	// Step 4: Combine Orders, Buyer, and Items
	var finalOrders []OrderDetailWithItemsResponse
	for _, order := range orders {
		finalOrders = append(finalOrders, OrderDetailWithItemsResponse{
			OrderResponse: OrderResponse{
				OrderID:     order.OrderID,
				TotalAmount: order.TotalAmount,
				Status:      order.Status,
				CreatedAt:   order.CreatedAt,
				UpdatedAt:   order.UpdatedAt,
			},
			Buyer: OrderBuyerResponse{
				ID:    order.BuyerID,
				Email: order.BuyerEmail,
				Phone: order.BuyerPhone,
				Name:  order.BuyerName,
			},
			Items: orderItemMap[order.OrderID],
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

	helper.SendPagination(c, http.StatusOK, "Orders retrieved successfully", finalOrders, pagination)
}

// GetOrderItemsDetail retrieves the details of a specific order for an authenticated user
// @Summary Get Order Items Detail
// @Description Retrieve details of a specific order, accessible only to the order's buyer
// @Tags Admin Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param order_id path int true "Order ID"
// @Success 200 {object} helper.SuccessResponse{data=OrderDetailWithItemsResponse} "Order details fetched successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid order ID"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 403 {object} helper.ErrorResponse "Access denied"
// @Failure 500 {object} helper.ErrorResponse "Failed to fetch order details"
// @Router /admin/order/{order_id} [get]
func GetOrderItemsDetail(c *gin.Context) {
	// Step 2: Parse Order ID from path parameter
	orderIDStr := c.Param("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Invalid order ID"})
		return
	}

	// Step 3: Verify ownership of the order
	var order models.Order
	err = config.DB.Where("id = ? ", orderID).First(&order).Error
	if err != nil {
		helper.SendError(c, http.StatusForbidden, []string{"You do not have access to this order"})
		return
	}
	type OrderResponseTmp struct {
		OrderID     uint64  `json:"order_id"`
		TotalAmount float64 `json:"total_amount"`
		Status      string  `json:"status"`
		CreatedAt   string  `json:"created_at"`
		UpdatedAt   string  `json:"updated_at"`
		BuyerID     uint64  `json:"buyer_id"`
		BuyerEmail  string  `json:"buyer_email"`
		BuyerName   string  `json:"buyer_name"`
		BuyerPhone  string  `json:"buyer_phone"`
	}

	// Step 4: Fetch Order Details
	var orderDetail OrderResponseTmp
	err = config.DB.Table("orders").
		Select(`
			orders.id AS order_id,
			COALESCE(profiles.name, '') AS buyer_name,
			orders.total_amount,
			orders.status,
			orders.created_at,
			orders.updated_at,
			profiles.name as buyer_name,
			users.id AS buyer_id,
			users.email AS buyer_email,
			users.phone AS buyer_phone`).
		Joins("JOIN profiles ON profiles.user_id = orders.buyer_id").
		Joins("JOIN users ON users.id = profiles.user_id").
		Order("orders.created_at DESC").
		Scan(&orderDetail).Error

	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to fetch order details"})
		return
	}

	// Step 5: Fetch Order Items Separately
	var orderItems []OrderItemResponse
	err = config.DB.Table("order_items").
		Select(`
			products.name AS product_name,
			order_items.quantity,
			order_items.price,
			(order_items.price * order_items.quantity) AS total_price,
			products.image_url`).
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("order_items.order_id = ?", orderID).
		Scan(&orderItems).Error

	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to fetch order items"})
		return
	}

	// Step 6: Attach items to order details
	orderDetail.CreatedAt = order.CreatedAt.Format(time.RFC3339)
	orderDetail.UpdatedAt = order.UpdatedAt.Format(time.RFC3339)

	finalResponse := OrderDetailWithItemsResponse{
		OrderResponse: OrderResponse{
			OrderID:     orderDetail.OrderID,
			TotalAmount: orderDetail.TotalAmount,
			Status:      orderDetail.Status,
			CreatedAt:   orderDetail.CreatedAt,
			UpdatedAt:   orderDetail.UpdatedAt,
		},
		Buyer: OrderBuyerResponse{
			ID:    orderDetail.BuyerID,
			Name:  orderDetail.BuyerName,
			Email: orderDetail.BuyerEmail,
			Phone: orderDetail.BuyerPhone,
		},
		Items: orderItems,
	}

	// Step 7: Send Response
	helper.SendSuccess(c, http.StatusOK, "Order details fetched successfully", finalResponse)
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
