package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Represents a product in the inventory
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// Represents a category for products
type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Mock data for products
var products = []Product{
	{ID: 1, Name: "Laptop", Price: 999.99, Stock: 10},
	{ID: 2, Name: "Smartphone", Price: 499.99, Stock: 25},
	{ID: 3, Name: "Tablet", Price: 299.99, Stock: 15},
	{ID: 4, Name: "Headphones", Price: 99.99, Stock: 60},
}

// Mock data for categories
var categories = []Category{
	{ID: 1, Name: "Electronics", Description: "Electronic devices and gadgets"},
	{ID: 2, Name: "Accessories", Description: "Related accessories and add-ons"},
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	fmt.Printf("Starting server on %s\n", addr)

	router := gin.Default()

	// Products
	router.GET("/api/products", handleGetProducts)
	router.POST("/api/products", handleCreateProduct)
	router.GET("/api/products/:id", handleGetProductByID)
	router.PUT("/api/products/:id", handleUpdateProduct)
	router.DELETE("/api/products/:id", handleDeleteProduct)

	// Categories
	router.GET("/categories", handleGetCategories)
	router.POST("/categories", handleCreateCategory)
	router.GET("/categories/:id", handleGetCategoryByID)
	router.PUT("/categories/:id", handleUpdateCategory)
	router.DELETE("/categories/:id", handleDeleteCategory)

	// Health check
	router.GET("/health", handleHealth)

	// Swagger UI and spec
	router.GET("/swagger/*any", handleSwagger)

	router.GET("/", handleRoot)
	err := router.Run(addr)
	if err != nil {
		panic(err)
	}
}

func handleGetProducts(c *gin.Context) {
	writeJSON(c, http.StatusOK, products)
}

func handleCreateProduct(c *gin.Context) {
	var newProduct Product
	if err := c.ShouldBindJSON(&newProduct); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	newProduct.ID = nextProductID()
	products = append(products, newProduct)
	writeJSON(c, http.StatusCreated, newProduct)
}

func handleGetProductByID(c *gin.Context) {
	id, ok := parseIDParam(c, "id", "Invalid product ID")
	if !ok {
		return
	}
	for _, product := range products {
		if product.ID == id {
			writeJSON(c, http.StatusOK, product)
			return
		}
	}
	writeError(c, http.StatusNotFound, "Product not found")
}

func handleUpdateProduct(c *gin.Context) {
	id, ok := parseIDParam(c, "id", "Invalid product ID")
	if !ok {
		return
	}
	var updated Product
	if err := c.ShouldBindJSON(&updated); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	for i, product := range products {
		if product.ID == id {
			updated.ID = product.ID
			products[i] = updated
			writeJSON(c, http.StatusOK, updated)
			return
		}
	}
	writeError(c, http.StatusNotFound, "Product not found")
}

func handleDeleteProduct(c *gin.Context) {
	id, ok := parseIDParam(c, "id", "Invalid product ID")
	if !ok {
		return
	}
	for i, product := range products {
		if product.ID == id {
			products = append(products[:i], products[i+1:]...)
			writeJSON(c, http.StatusOK, map[string]string{"message": "Product deleted"})
			return
		}
	}
	writeError(c, http.StatusNotFound, "Product not found")
}

func handleGetCategories(c *gin.Context) {
	writeJSON(c, http.StatusOK, categories)
}

func handleCreateCategory(c *gin.Context) {
	var newCategory Category
	if err := c.ShouldBindJSON(&newCategory); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	newCategory.ID = nextCategoryID()
	categories = append(categories, newCategory)
	writeJSON(c, http.StatusCreated, newCategory)
}

func handleGetCategoryByID(c *gin.Context) {
	id, ok := parseIDParam(c, "id", "Invalid category ID")
	if !ok {
		return
	}
	for _, category := range categories {
		if category.ID == id {
			writeJSON(c, http.StatusOK, category)
			return
		}
	}
	writeError(c, http.StatusNotFound, "Category not found")
}

func handleUpdateCategory(c *gin.Context) {
	id, ok := parseIDParam(c, "id", "Invalid category ID")
	if !ok {
		return
	}
	var updated Category
	if err := c.ShouldBindJSON(&updated); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	for i, category := range categories {
		if category.ID == id {
			updated.ID = category.ID
			categories[i] = updated
			writeJSON(c, http.StatusOK, updated)
			return
		}
	}
	writeError(c, http.StatusNotFound, "Category not found")
}

func handleDeleteCategory(c *gin.Context) {
	id, ok := parseIDParam(c, "id", "Invalid category ID")
	if !ok {
		return
	}
	for i, category := range categories {
		if category.ID == id {
			categories = append(categories[:i], categories[i+1:]...)
			writeJSON(c, http.StatusOK, map[string]string{"message": "Category deleted"})
			return
		}
	}
	writeError(c, http.StatusNotFound, "Category not found")
}

func handleRoot(c *gin.Context) {
	writeJSON(c, http.StatusOK, map[string]string{"message": "Hello, Ini Backend Program Kasir!"})
}

func handleHealth(c *gin.Context) {
	writeJSON(c, http.StatusOK, map[string]string{"status": "ok", "message": "Service is running"})
}

func handleSwagger(c *gin.Context) {
	if c.Param("any") == "/doc.json" {
		c.Data(http.StatusOK, "application/json", []byte(swaggerSpec))
		return
	}
	httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json"))(c.Writer, c.Request)
}

func writeJSON(c *gin.Context, status int, payload interface{}) {
	c.JSON(status, payload)
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func nextCategoryID() int {
	maxID := 0
	for _, category := range categories {
		if category.ID > maxID {
			maxID = category.ID
		}
	}
	return maxID + 1
}

func nextProductID() int {
	maxID := 0
	for _, product := range products {
		if product.ID > maxID {
			maxID = product.ID
		}
	}
	return maxID + 1
}

func parseIDParam(c *gin.Context, name, errMsg string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil {
		writeError(c, http.StatusBadRequest, errMsg)
		return 0, false
	}
	return id, true
}

const swaggerSpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Kasir API",
    "version": "1.0.0"
  },
  "paths": {
    "/health": {
      "get": {
        "responses": {
          "200": { "description": "OK" }
        }
      }
    },
    "/api/products": {
      "get": {
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": { "$ref": "#/components/schemas/Product" }
                }
              }
            }
          }
        }
      },
      "post": {
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/Product" }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Created",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Product" }
              }
            }
          }
        }
      }
    },
    "/api/products/{id}": {
      "get": {
        "parameters": [ { "$ref": "#/components/parameters/IdParam" } ],
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Product" }
              }
            }
          },
          "404": { "description": "Not Found" }
        }
      },
      "put": {
        "parameters": [ { "$ref": "#/components/parameters/IdParam" } ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/Product" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Product" }
              }
            }
          },
          "404": { "description": "Not Found" }
        }
      },
      "delete": {
        "parameters": [ { "$ref": "#/components/parameters/IdParam" } ],
        "responses": {
          "200": { "description": "OK" },
          "404": { "description": "Not Found" }
        }
      }
    },
    "/categories": {
      "get": {
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": { "$ref": "#/components/schemas/Category" }
                }
              }
            }
          }
        }
      },
      "post": {
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/Category" }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Created",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Category" }
              }
            }
          }
        }
      }
    },
    "/categories/{id}": {
      "get": {
        "parameters": [ { "$ref": "#/components/parameters/IdParam" } ],
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Category" }
              }
            }
          },
          "404": { "description": "Not Found" }
        }
      },
      "put": {
        "parameters": [ { "$ref": "#/components/parameters/IdParam" } ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/Category" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Category" }
              }
            }
          },
          "404": { "description": "Not Found" }
        }
      },
      "delete": {
        "parameters": [ { "$ref": "#/components/parameters/IdParam" } ],
        "responses": {
          "200": { "description": "OK" },
          "404": { "description": "Not Found" }
        }
      }
    }
  },
  "components": {
    "parameters": {
      "IdParam": {
        "name": "id",
        "in": "path",
        "required": true,
        "schema": { "type": "integer" }
      }
    },
    "schemas": {
      "Product": {
        "type": "object",
        "properties": {
          "id": { "type": "integer" },
          "name": { "type": "string" },
          "price": { "type": "number", "format": "float" },
          "stock": { "type": "integer" }
        }
      },
      "Category": {
        "type": "object",
        "properties": {
          "id": { "type": "integer" },
          "name": { "type": "string" },
          "description": { "type": "string" }
        }
      }
    }
  }
}`
