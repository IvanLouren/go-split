// @title           GoSplit API
// @version         1.0
// @description     A Splitwise-like expense splitting REST API.

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/IvanLouren/GoSplit/docs"
	"github.com/IvanLouren/GoSplit/internal/auth"
	"github.com/IvanLouren/GoSplit/internal/balances"
	"github.com/IvanLouren/GoSplit/internal/expenses"
	"github.com/IvanLouren/GoSplit/internal/groups"
	"github.com/IvanLouren/GoSplit/internal/settlements"
	"github.com/IvanLouren/GoSplit/pkg/database"
	"github.com/IvanLouren/GoSplit/pkg/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {

	godotenv.Load()

	database.Connect()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// init auths
	authService := auth.NewService(database.DB)
	authHandler := auth.NewHandler(authService)

	// init groups
	groupService := groups.NewService(database.DB)
	groupHandler := groups.NewHandler(groupService)

	// init expenses
	expenseService := expenses.NewService(database.DB)
	expenseHandler := expenses.NewHandler(expenseService)

	// init settlements
	settlementService := settlements.NewService(database.DB)
	settlementHandler := settlements.NewHandler(settlementService)

	// init balances
	balanceService := balances.NewService(database.DB)
	balanceHandler := balances.NewHandler(balanceService)

	// auth routes
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)

	// group routes
	mux.Handle("POST /api/groups", middleware.AuthRequired(http.HandlerFunc(groupHandler.CreateGroup)))
	mux.Handle("GET /api/groups", middleware.AuthRequired(http.HandlerFunc(groupHandler.GetGroups)))
	mux.Handle("GET /api/groups/{id}", middleware.AuthRequired(http.HandlerFunc(groupHandler.GetGroup)))
	mux.Handle("PUT /api/groups/{id}", middleware.AuthRequired(http.HandlerFunc(groupHandler.UpdateGroup)))
	mux.Handle("DELETE /api/groups/{id}", middleware.AuthRequired(http.HandlerFunc(groupHandler.DeleteGroup)))
	mux.Handle("POST /api/groups/{id}/members", middleware.AuthRequired(http.HandlerFunc(groupHandler.AddMember)))
	mux.Handle("DELETE /api/groups/{id}/members/{user_id}", middleware.AuthRequired(http.HandlerFunc(groupHandler.RemoveMember)))

	// expense routes
	mux.Handle("POST /api/groups/{id}/expenses", middleware.AuthRequired(http.HandlerFunc(expenseHandler.CreateExpense)))
	mux.Handle("GET /api/groups/{id}/expenses", middleware.AuthRequired(http.HandlerFunc(expenseHandler.GetExpenses)))
	mux.Handle("GET /api/groups/{id}/expenses/{expenseId}", middleware.AuthRequired(http.HandlerFunc(expenseHandler.GetExpense)))
	mux.Handle("DELETE /api/groups/{id}/expenses/{expenseId}", middleware.AuthRequired(http.HandlerFunc(expenseHandler.DeleteExpense)))

	// settlement routes
	mux.Handle("POST /api/groups/{id}/settlements", middleware.AuthRequired(http.HandlerFunc(settlementHandler.CreateSettlement)))
	mux.Handle("GET /api/groups/{id}/settlements", middleware.AuthRequired(http.HandlerFunc(settlementHandler.GetSettlements)))

	// balance routes
	mux.Handle("GET /api/groups/{id}/balances", middleware.AuthRequired(http.HandlerFunc(balanceHandler.GetBalances)))

	// swagger UI
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("server starting on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
