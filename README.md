# 🛠️ **Deketna Backend (BE)**

Backend service for **Deketna**, an application for managing product stock, orders, and customer interactions.

---

## 🚀 **Project Overview**

- **Language:** Golang
- **Framework:** Gin
- **Database:** PostgreSQL
- **Cache:** Redis (planned)
- **Authentication:** JWT (JSON Web Token)
- **Pagination:** Includes metadata (`totalItem`, `isNext`, `isPrev`)

This backend powers the **Deketna** frontend, ensuring fast, secure, and scalable communication between the client and server.

---

## 📚 **Features**

1. **Authentication System:**

   - User Registration & Login (JWT-based)
   - Secure password hashing

2. **Order Management:**

   - View, Create, and Update Orders
   - Detailed order history with buyer and product details

3. **Pagination Support:**

   - Metadata for `totalItems`, `totalPages`, `isNext`, `isPrev`

4. **Product Management:**

   - Add, Update, and Remove Products
   - Generate QR codes for product tracking

5. **Scalability:**
   - Optimized query performance
   - Ready for Redis caching integration

---

## 🏗️ **Project Structure**

```plaintext
📂 deketna-backend
├── 📁 config          # Database & app configurations
├── 📁 handlers        # Route Handlers
│   ├── 📁 admin       # Admin-specific endpoints
│   ├── 📁 auth        # Authentication endpoints
│   └── 📁 buyer       # Buyer-specific endpoints
├── 📁 models         # Database models
├── 📁 helper         # Utility functions
├── 📁 routes         # API route definitions
├── 📁 middleware     # JWT & Request Validation Middleware
├── 📁 dto           # Data Transfer Objects (Request/Response Structures)
└── main.go          # Application Entry Point
```

---

## ⚙️ **Installation & Setup**

### **1. Clone the Repository**

```bash
git clone https://github.com/yourusername/deketna-backend.git
cd deketna-backend
```

### **2. Environment Variables**

Create a `.env` file in the root directory:

```env
DB_HOST==*****
DB_USER==*****
DB_PASSWORD==*****
DB_NAME==*****
DB_PORT==*****
DB_SSLMODE==*****
SUPABASE_URL=****
SUPABASE_API_KEY=****
SUPABASE_BUCKET=*****
JWT_SECRET==*****
```

### **3. Install Dependencies**

```bash
go mod tidy
```

### **4. Run Migrations**

Ensure your database is set up:

```bash
go run main.go migrate
```

### **5. Start the Server**

```bash
go run main.go
```

Server will run on `http://localhost:8080`

---

## 📑 **API Documentation**

### **Authentication**

- `POST /auth/register` — Register a new user
- `POST /auth/login` — User login

### **Orders (Admin)**

- `GET /admin/orders` — View all orders with pagination
- `GET /admin/orders/:id` — Get order details by ID

### **Products**

- `POST /admin/products` — Add new product
- `GET /products` — List all products

### **Pagination Example Response:**

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "limit": 10,
    "totalItems": 100,
    "totalPages": 10,
    "isNext": true,
    "isPrev": false
  }
}
```

For a full list of APIs, refer to **API Documentation** using tools like **Swagger** or **Postman**.

---

## 🔒 **Security**

- JWT Authentication for API routes
- Environment variables for sensitive data
- Password hashing with bcrypt

---

## 🧪 **Testing**

Run tests:

```bash
go test ./...
```

---

## 📊 **Future Improvements**

- Integration with Redis for caching
- Real-time notifications using WebSockets
- Enhanced reporting and analytics

---

## 🤝 **Contributing**

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Commit changes: `git commit -m "Add new feature"`
4. Push to branch: `git push origin feature/new-feature`
5. Create a Pull Request

---

## 📝 **License**

This project is licensed under the **MIT License**.

---

## 📬 **Contact**

- **Email:** feriteja@gmail.com
- **GitHub:** [feriteja](https://github.com/feriteja)

---

Happy coding! 🚀✨
