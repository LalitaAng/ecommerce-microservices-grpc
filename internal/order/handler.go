package order

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service *OrderService
}

func NewOrderHandler(service *OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(PageDefault)))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(PageSizeDefault)))

	orders, totalCount, err := h.service.ListOrders(c.Request.Context(), status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	c.JSON(http.StatusOK, gin.H{
		"orders":      orders,
		"total_count": totalCount,
		"total_pages": totalPages,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var orderRequest CreateOrderRequest
	if err := c.ShouldBindJSON(&orderRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), userID.(string), orderRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) PayOrder(c *gin.Context) {
    orderID := c.Param("order_id")
    if orderID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "order ID is required"})
        return
    }

    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    order, err := h.service.PayOrder(c.Request.Context(), orderID, userID.(string))
    if err != nil {
        switch err.Error() {
        case "order is already paid":
            c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
        case "order does not belong to user":
            c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        case "cannot pay for a cancelled order":
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payment"})
        }
        return
    }

    c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetOrderDetails(c *gin.Context) {
	orderID := c.Param("order_id")

	order, err := h.service.GetOrderDetails(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("order_id")

	order, err := h.service.CancelOrder(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order":   order,
		"message": "Order cancelled successfully",
	})
}

func (h *OrderHandler) ListOrderProducts(c *gin.Context) {
	orderID := c.Param("order_id")

	items, err := h.service.ListOrderProducts(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"items":    items,
	})
}

func (h *OrderHandler) RefundOrder(c *gin.Context) {
    orderID := c.Param("order_id")
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    order, err := h.service.RefundOrder(c.Request.Context(), orderID, userID.(string))
    if err != nil {
        switch err.Error() {
        case "order does not belong to user":
            c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        case "only completed payments can be refunded":
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process refund"})
        }
        return
    }

    c.JSON(http.StatusOK, order)
}
