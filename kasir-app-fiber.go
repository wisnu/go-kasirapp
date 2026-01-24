package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
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

	app := fiber.New()

	// Products
	app.Get("/api/products", handleGetProducts)
	app.Post("/api/products", handleCreateProduct)
	app.Get("/api/products/:id", handleGetProductByID)
	app.Put("/api/products/:id", handleUpdateProduct)
	app.Delete("/api/products/:id", handleDeleteProduct)

	// Categories
	app.Get("/categories", handleGetCategories)
	app.Post("/categories", handleCreateCategory)
	app.Get("/categories/:id", handleGetCategoryByID)
	app.Put("/categories/:id", handleUpdateCategory)
	app.Delete("/categories/:id", handleDeleteCategory)

	// Health check
	app.Get("/health", handleHealth)

	// Swagger UI and spec
	app.Get("/swagger", handleSwaggerUI)
	app.Get("/swagger/", handleSwaggerUI)
	app.Get("/swagger/index.html", handleSwaggerUI)
	app.Get("/swagger/doc.json", handleSwaggerDoc)

	app.Get("/", handleRoot)
	err := app.Listen(addr)
	if err != nil {
		panic(err)
	}
}

func handleGetProducts(c *fiber.Ctx) error {
	return writeJSON(c, http.StatusOK, products)
}

func handleCreateProduct(c *fiber.Ctx) error {
	var newProduct Product
	if err := c.BodyParser(&newProduct); err != nil {
		return writeError(c, http.StatusBadRequest, err.Error())
	}
	newProduct.ID = nextProductID()
	products = append(products, newProduct)
	return writeJSON(c, http.StatusCreated, newProduct)
}

func handleGetProductByID(c *fiber.Ctx) error {
	id, ok := parseIDParam(c, "id", "Invalid product ID")
	if !ok {
		return nil
	}
	for _, product := range products {
		if product.ID == id {
			return writeJSON(c, http.StatusOK, product)
		}
	}
	return writeError(c, http.StatusNotFound, "Product not found")
}

func handleUpdateProduct(c *fiber.Ctx) error {
	id, ok := parseIDParam(c, "id", "Invalid product ID")
	if !ok {
		return nil
	}
	var updated Product
	if err := c.BodyParser(&updated); err != nil {
		return writeError(c, http.StatusBadRequest, err.Error())
	}
	for i, product := range products {
		if product.ID == id {
			updated.ID = product.ID
			products[i] = updated
			return writeJSON(c, http.StatusOK, updated)
		}
	}
	return writeError(c, http.StatusNotFound, "Product not found")
}

func handleDeleteProduct(c *fiber.Ctx) error {
	id, ok := parseIDParam(c, "id", "Invalid product ID")
	if !ok {
		return nil
	}
	for i, product := range products {
		if product.ID == id {
			products = append(products[:i], products[i+1:]...)
			return writeJSON(c, http.StatusOK, map[string]string{"message": "Product deleted"})
		}
	}
	return writeError(c, http.StatusNotFound, "Product not found")
}

func handleGetCategories(c *fiber.Ctx) error {
	return writeJSON(c, http.StatusOK, categories)
}

func handleCreateCategory(c *fiber.Ctx) error {
	var newCategory Category
	if err := c.BodyParser(&newCategory); err != nil {
		return writeError(c, http.StatusBadRequest, err.Error())
	}
	newCategory.ID = nextCategoryID()
	categories = append(categories, newCategory)
	return writeJSON(c, http.StatusCreated, newCategory)
}

func handleGetCategoryByID(c *fiber.Ctx) error {
	id, ok := parseIDParam(c, "id", "Invalid category ID")
	if !ok {
		return nil
	}
	for _, category := range categories {
		if category.ID == id {
			return writeJSON(c, http.StatusOK, category)
		}
	}
	return writeError(c, http.StatusNotFound, "Category not found")
}

func handleUpdateCategory(c *fiber.Ctx) error {
	id, ok := parseIDParam(c, "id", "Invalid category ID")
	if !ok {
		return nil
	}
	var updated Category
	if err := c.BodyParser(&updated); err != nil {
		return writeError(c, http.StatusBadRequest, err.Error())
	}
	for i, category := range categories {
		if category.ID == id {
			updated.ID = category.ID
			categories[i] = updated
			return writeJSON(c, http.StatusOK, updated)
		}
	}
	return writeError(c, http.StatusNotFound, "Category not found")
}

func handleDeleteCategory(c *fiber.Ctx) error {
	id, ok := parseIDParam(c, "id", "Invalid category ID")
	if !ok {
		return nil
	}
	for i, category := range categories {
		if category.ID == id {
			categories = append(categories[:i], categories[i+1:]...)
			return writeJSON(c, http.StatusOK, map[string]string{"message": "Category deleted"})
		}
	}
	return writeError(c, http.StatusNotFound, "Category not found")
}

func handleRoot(c *fiber.Ctx) error {
	return writeJSON(c, http.StatusOK, map[string]string{"message": "Hello, Ini Backend Program Kasir!"})
}

func handleHealth(c *fiber.Ctx) error {
	return writeJSON(c, http.StatusOK, map[string]string{"status": "ok", "message": "Service is running"})
}

func handleSwaggerUI(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Status(http.StatusOK).SendString(swaggerUIHTML)
}

func handleSwaggerDoc(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	return c.Status(http.StatusOK).SendString(swaggerSpec)
}

func writeJSON(c *fiber.Ctx, status int, payload interface{}) error {
	return c.Status(status).JSON(payload)
}

func writeError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
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

func parseIDParam(c *fiber.Ctx, name, errMsg string) (int, bool) {
	id, err := strconv.Atoi(c.Params(name))
	if err != nil {
		_ = writeError(c, http.StatusBadRequest, errMsg)
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

const swaggerUIHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Kasir API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = function () {
        SwaggerUIBundle({
          url: "/swagger/doc.json",
          dom_id: "#swagger-ui"
        });
      };
    </script>
  </body>
</html>`
