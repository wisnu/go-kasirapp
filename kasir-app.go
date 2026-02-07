package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"kasir-api/config"
	"kasir-api/database"
	"kasir-api/handlers"
	"kasir-api/repositories"
	"kasir-api/services"
)

//go:embed openapi.yaml
var openapiSpec embed.FS

func main() {
	// Load configuration
	cfg, cfgErr := config.LoadConfig()
	if cfgErr != nil {
		panic(cfgErr)
	}

	db, db_err := database.Connect(cfg.DB)
	if db_err != nil {
		panic(db_err)
	}
	defer db.Close()

	// Wire: repository -> service -> handler
	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	categoryRepo := repositories.NewCategoryRepository(db)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	port := cfg.App.Port
	addr := ":" + strconv.Itoa(port)
	fmt.Printf("Starting server on %s\n", addr)

	// Handle API routes
	http.HandleFunc("/api/products", productHandler.Handle)
	http.HandleFunc("/api/products/", productHandler.Handle)

	http.HandleFunc("/categories", categoryHandler.Handle)
	http.HandleFunc("/categories/", categoryHandler.Handle)

	// API docs (Scalar)
	http.HandleFunc("/docs", handleDocs)
	http.HandleFunc("/docs/openapi.yaml", handleOpenAPISpec)

	// Health check
	http.HandleFunc("/health", handleHealth)

	http.HandleFunc("/", handleRoot)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Kasir API – Docs</title>
  </head>
  <body>
    <div id="app"></div>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>
      Scalar.createApiReference('#app', {
        url: '/docs/openapi.yaml',
      })
    </script>
  </body>
</html>`)
}

func handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec, err := openapiSpec.ReadFile("openapi.yaml")
	if err != nil {
		http.Error(w, "failed to read spec", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.Write(spec)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "Hello, Ini Backend Program Kasir!"})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "Service is running"})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
