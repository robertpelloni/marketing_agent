package main
import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)
func main() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()
	client := crm.NewRestCRMClient(server.URL, "test-key")
	err := client.PushDeal(context.Background(), db.Deal{ID: 1}, db.Company{Name: "Test"}, "test")
	if err != nil { log.Fatalf("Failed: %v", err) }
	fmt.Println("CRM Integration verified")
}
