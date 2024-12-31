package admin

import (
	"deketna/config"
	"deketna/helper"
	"deketna/models"
	"deketna/utils"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// @Summary Add a product
// @Description Admin adds a new product
// @Tags Admin Product
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Product Name"
// @Param price formData number true "Product Price"
// @Param stock formData integer true "Product Stock"
// @Param category_id formData integer true "Product Category"
// @Param image formData file true "Product Image"
// @Success 201 {object} helper.SuccessResponse "Product"
// @Success 400 {object} helper.ErrorResponse "Validation Error"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 403 {object} helper.ErrorResponse "Access forbidden"
// @Router /admin/product [post]
func AddProduct(c *gin.Context) {
	// Parse form data
	var req AddProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Retrieve admin ID from JWT token
	claims := c.MustGet("claims").(jwt.MapClaims)
	adminID, exists := claims["userid"].(float64)
	if !exists {
		helper.SendError(c, http.StatusInternalServerError, []string{"User ID not found in claims"})
		return
	}

	// Handle file upload
	file, err := c.FormFile("image")
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Failed to upload image"})
		return
	}

	// Ensure the tmp folder exists
	tmpDir := "tmp"
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		err = os.Mkdir(tmpDir, os.ModePerm)
		if err != nil {
			helper.SendError(c, http.StatusInternalServerError, []string{"Failed to create tmp directory"})
			return
		}
	}

	// Sanitize the filename
	safeFilename := filepath.Base(file.Filename) // Prevent path traversal
	tempFilePath := filepath.Join(tmpDir, safeFilename)

	// Save the uploaded file temporarily
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Failed to save temporary image"})
		return
	}

	// Upload to Supabase
	imageURL, err := utils.UploadImageToSupabase(tempFilePath, safeFilename)
	fmt.Println("imageURL:", imageURL)
	if err != nil {
		// Clean up the temporary file even on upload failure
		os.Remove(tempFilePath)
		helper.SendError(c, http.StatusBadRequest, []string{fmt.Sprintf("Failed to upload image: %v", err)})
		return
	}

	// Remove the temporary file after successful upload
	err = os.Remove(tempFilePath)
	if err != nil {
		fmt.Println("Warning: Failed to delete temporary file:", tempFilePath)
	}

	// Create product record in DB
	product := models.Product{
		Name:     req.Name,
		Price:    req.Price,
		Stock:    req.Stock,
		SellerID: uint64(adminID),
		ImageURL: imageURL,
	}

	if err := config.DB.Create(&product).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to add product to database"})
		return
	}

	// Respond with success
	helper.SendSuccess(c, http.StatusCreated, "Product added successfully", product)
}

// GetProducts retrieves a paginated list of products with seller details
// @Summary Get Products
// @Description Retrieve a paginated list of products with seller details
// @Tags   Admin Product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 25)"
// @Param seller_id query int false "id of seller (default: 1)"
// @Param seller_name query string false "Name of seller (default: Deketna)"
// @Param product_name query string false "Name of product (default: botol)"
// @Success 200 {object} helper.PaginationResponse{data=[]GetProductResponseComplete} "List of products with seller details"
// @Failure 400 {object} helper.ErrorResponse "Invalid query parameters"
// @Router /admin/products [get]
func GetProduct(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))

	sellerIDParam := c.Query("seller_id")
	var sellerID *uint64
	if sellerIDParam != "" {
		id, err := strconv.ParseUint(sellerIDParam, 10, 64)
		if err == nil {
			sellerID = &id
		}
	}

	sellerName := c.Query("seller_name")
	var sellerNamePtr *string
	if sellerName != "" {
		sellerNamePtr = &sellerName
	}

	productName := c.Query("product_name")
	var productNamePtr *string
	if productName != "" {
		productNamePtr = &productName
	}

	products, totalItems, err := _getProductsPaginated(config.DB, page, limit, sellerID, sellerNamePtr, productNamePtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	totalPages := (int(totalItems) + limit - 1) / limit
	pagination := helper.PaginationMetadata{
		Page:       page,
		Limit:      limit,
		TotalItems: int(totalItems),
		TotalPages: totalPages,
		IsNext:     page < totalPages,
		IsPrev:     page > 1,
	}

	helper.SendPagination(c, http.StatusOK, "Products retrieved successfully", products, pagination)

}

// GetProducts retrieves a detail of product with seller
// @Summary Get Product Detail
// @Description Retrieve a detail of products with seller
// @Tags   Admin Product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} helper.SuccessResponse{data=GetProductResponseComplete} "List of products with seller details"
// @Failure 400 {object} helper.ErrorResponse "Invalid query parameters"
// @Router /admin/product/{id} [get]
func GetProductDetail(c *gin.Context) {
	productId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	product, err := _getProductDetail(config.DB, productId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	helper.SendSuccess(c, http.StatusOK, "Products retrieved successfully", product)

}

// @Summary Edit a product
// @Description Admin edit a product
// @Tags Admin Product
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param name formData string false "Product Name"
// @Param price formData number false "Product Price"
// @Param stock formData integer false "Product Stock"
// @Param category_id formData integer false "Product Category"
// @Param image formData file false "Product Image"
// @Success 200 {object} helper.SuccessResponse "Product"
// @Failure 400 {object} helper.ErrorResponse "Validation Error"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 403 {object} helper.ErrorResponse "Access forbidden"
// @Router /admin/product/{id} [put]
func AdminEditProduct(c *gin.Context) {
	// Parse Product ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Invalid product ID"})
		return
	}

	// Parse Optional Form Parameters
	var req ProductEditRequest

	req.Name = c.PostForm("name")
	if price := c.PostForm("price"); price != "" {
		if p, err := strconv.ParseFloat(price, 64); err == nil {
			req.Price = &p
		}
	}

	if stock := c.PostForm("stock"); stock != "" {
		if s, err := strconv.Atoi(stock); err == nil {
			req.Stock = &s
		}
	}

	if categoryID := c.PostForm("category_id"); categoryID != "" {
		if cid, err := strconv.ParseUint(categoryID, 10, 64); err == nil {
			cidUint := uint(cid)
			req.CategoryID = &cidUint
		}
	}

	// Handle Optional Image Upload
	var imageURL *string
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		imageURL, err = _handleImageUpload(c, file)
		if err != nil {
			helper.SendError(c, http.StatusInternalServerError, []string{err.Error()})
			return
		}
		req.ImageURL = imageURL
	}

	// Perform Product Update
	product, err := _editProduct(config.DB, id, req)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to update product"})
		return
	}

	// Success Response
	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"product": product,
	})
}

// @Summary Delete a product
// @Description Admin delete a product
// @Tags Admin Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} helper.SuccessResponse "Product"
// @Success 400 {object} helper.ErrorResponse "Validation Error"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 403 {object} helper.ErrorResponse "Access forbidden"
// @Router /admin/product/{id} [delete]
func AdminDeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	if err := _deleteProduct(config.DB, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func _getProductsPaginated(db *gorm.DB, page, limit int, sellerID *uint64, sellerName *string, productName *string) ([]GetProductResponseComplete, int64, error) {
	var products []GetProductResponseComplete
	var totalItems int64

	// Calculate offset
	offset := (page - 1) * limit

	query := db.Model(&models.Product{}).
		Preload("Seller").
		Preload("Category")

	if sellerID != nil {
		query = query.Where("products.seller_id = ?", *sellerID)
	}

	if sellerName != nil && *sellerName != "" {
		query = query.Joins("JOIN profiles ON profiles.user_id = products.seller_id").
			Where("LOWER(profiles.name) ILIKE LOWER(?)", "%"+*sellerName+"%")
	}

	if productName != nil && *productName != "" {
		query = query.Where("LOWER(name) ILIKE LOWER(?)", "%"+*productName+"%")
	}

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	userJSON, _ := json.MarshalIndent(products, "", "  ")
	fmt.Println(string(userJSON))

	// Map to DTO
	var response = make([]GetProductResponseComplete, len(products))

	for i, product := range products {
		response[i] = GetProductResponseComplete{
			GetProductResponse: GetProductResponse{
				ID:        product.ID,
				Name:      product.Name,
				Price:     product.Price,
				Stock:     product.Stock,
				SellerID:  product.SellerID,
				ImageURL:  product.ImageURL,
				CreatedAt: product.CreatedAt,
				UpdatedAt: product.UpdatedAt,
			},
			Seller: Profile{
				ID:       product.Seller.ID,
				Name:     product.Seller.Name,     // Adjust if `Name` comes from profiles
				ImageURL: product.Seller.ImageURL, // Profiles image must be handled manually
			},
			Category: Category{
				ID:          product.Category.ID,
				Name:        product.Category.Name,
				Description: product.Category.Description,
			},
		}
	}

	return products, totalItems, nil
}
func _getProductDetail(db *gorm.DB, productId uint64) (*GetProductResponseComplete, error) {

	var product GetProductResponseComplete

	query := db.Model(&models.Product{}).
		Preload("Seller").
		Preload("Category")

	if err := query.First(&product, productId).Error; err != nil {
		return nil, err
	}

	userJSON, _ := json.MarshalIndent(product, "", "  ")
	fmt.Println(string(userJSON))

	// Map to DTO

	response := GetProductResponseComplete{
		GetProductResponse: GetProductResponse{
			ID:        product.ID,
			Name:      product.Name,
			Price:     product.Price,
			Stock:     product.Stock,
			SellerID:  product.SellerID,
			ImageURL:  product.ImageURL,
			CreatedAt: product.CreatedAt,
			UpdatedAt: product.UpdatedAt,
		},
		Seller: Profile{
			ID:       product.Seller.ID,
			Name:     product.Seller.Name,     // Adjust if `Name` comes from profiles
			ImageURL: product.Seller.ImageURL, // Profiles image must be handled manually
		},
		Category: Category{
			ID:          product.Category.ID,
			Name:        product.Category.Name,
			Description: product.Category.Description,
		},
	}

	return &response, nil
}

func _editProduct(db *gorm.DB, id uint64, req ProductEditRequest) (*GetProductResponse, error) {
	var product models.Product

	// Find product by ID
	if err := db.Omit("Seller", "Category").First(&product, id).Error; err != nil {
		return nil, err
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}
	if req.ImageURL != nil {
		product.ImageURL = *req.ImageURL
	}

	if err := db.Omit("Seller", "Category").Save(&product).Error; err != nil {
		return nil, err
	}

	// Map to DTO
	response := GetProductResponse{
		ID:         product.ID,
		Name:       product.Name,
		Price:      product.Price,
		Stock:      product.Stock,
		SellerID:   product.SellerID,
		CategoryID: product.CategoryID,
		ImageURL:   product.ImageURL,
		CreatedAt:  product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  product.UpdatedAt.Format(time.RFC3339),
	}

	return &response, nil
}

func _deleteProduct(db *gorm.DB, id uint64) error {
	// Delete product by ID
	if err := db.Delete(&models.Product{}, id).Error; err != nil {
		return err
	}
	return nil
}

func _handleImageUpload(c *gin.Context, file *multipart.FileHeader) (*string, error) {
	// Ensure the tmp folder exists
	tmpDir := "tmp"
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		err = os.Mkdir(tmpDir, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("failed to create tmp directory: %v", err)
		}
	}

	// Sanitize the filename to prevent path traversal
	safeFilename := filepath.Base(file.Filename)
	tempFilePath := filepath.Join(tmpDir, safeFilename)

	// Save the uploaded file to the tmp directory
	err := c.SaveUploadedFile(file, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to save temporary image: %v", err)
	}

	// Upload to Supabase
	imageURL, err := utils.UploadImageToSupabase(tempFilePath, safeFilename)
	if err != nil {
		// Clean up temporary file even on upload failure
		os.Remove(tempFilePath)
		return nil, fmt.Errorf("failed to upload image to Supabase: %v", err)
	}

	// Remove the temporary file after successful upload
	if err := os.Remove(tempFilePath); err != nil {
		fmt.Println("Warning: Failed to delete temporary file:", tempFilePath)
	}

	return &imageURL, nil
}
