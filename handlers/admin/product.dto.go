package admin

type AddProductRequest struct {
	Name  string  `form:"name" binding:"required"`
	Price float64 `form:"price" binding:"required,gt=0"`
	Stock int     `form:"stock" binding:"required,gt=0"`
}
