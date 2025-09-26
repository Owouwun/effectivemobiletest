package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Owouwun/effectivemobiletest/internal/core/logic/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const dateLayout = "01-2006"

type SubscriptionService interface {
	CreateService(ctx context.Context, srv *services.Service) error
	GetService(ctx context.Context, ID uuid.UUID) (*services.Service, error)
	GetServices(ctx context.Context) ([]*services.Service, error)
	UpdateService(ctx context.Context, srv *services.Service) error
	DeleteService(ctx context.Context, ID uuid.UUID) error
	CumulateServices(ctx context.Context, filters *services.Filters) (int, error)
}

type SubscriptionHandler struct {
	subscriptionService SubscriptionService
}

func NewSubscriptionHandler(ss SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: ss,
	}
}

type CreateRequest struct {
	ServiceName string    `json:"service_name" example:"My Service"`
	Price       int       `json:"price" example:"500"`
	UserID      uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	StartDate   string    `json:"start_date" example:"01-2024"`
	EndDate     *string   `json:"end_date,omitempty" example:"12-2025"`
}

// @BasePath /service

// CreateService godoc
// @Summary      Create a service
// @Description  Create a service instance
// @Tags         services
// @Param    	 request  body  CreateRequest  true  "Create Service Request"
// @Accept       json
// @Produce      json
// @Success      201  {object}  services.Service
// @Failure      400  {object}  map[string]any "Invalid request body or date format"
// @Failure      500  {object}  map[string]any "Internal server error"
// @Router       / [post]
func (h *SubscriptionHandler) CreateService(c *gin.Context) {
	var req *CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Invalid request body")
		return
	}

	startDate, err := time.Parse(dateLayout, req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("invalid start date")
		return
	}

	srv := &services.Service{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
	}

	if req.EndDate != nil {
		endDate, err := time.Parse(dateLayout, *req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date", "details": err.Error()})
			logrus.WithFields(logrus.Fields{
				"path":    c.Request.URL.Path,
				"details": err.Error(),
			}).Warn("invalid end date")
			return
		}

		srv.EndDate = &endDate
	}

	if err := h.subscriptionService.CreateService(c, srv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add new service", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("failed to add new service")
		return
	}

	c.JSON(http.StatusCreated, srv)
	logrus.WithFields(logrus.Fields{
		"path":   c.Request.URL.Path,
		"status": http.StatusCreated,
		"body":   srv,
	}).Debug("Success respond")
}

// GetService godoc
// @Summary      Get a service by name
// @Description  Retrieves a single service instance by its unique name.
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        ID  path  string  true  "UUID of the service to retrieve" example:"123e4567-e89b-12d3-a456-426614174009"
// @Success      200  {object}  services.Service  "Successfully retrieved service"
// @Failure      400  {object}  map[string]any    "Invalid UUID"
// @Failure      404  {object}  map[string]any    "Service not found"
// @Failure      500  {object}  map[string]any    "Internal server error"
// @Router       /{id} [get]
func (h *SubscriptionHandler) GetService(c *gin.Context) {
	ID, err := uuid.Parse(c.Param("ID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Invalid UUID")
		return
	}

	srv, err := h.subscriptionService.GetService(c, ID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			logrus.WithFields(logrus.Fields{
				"path":    c.Request.URL.Path,
				"details": err.Error(),
			}).Warn("Service not found")
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get service", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Failed to get service")
		return
	}

	c.JSON(http.StatusOK, srv)
	logrus.WithFields(logrus.Fields{
		"path":   c.Request.URL.Path,
		"status": http.StatusOK,
		"body":   srv,
	}).Debug("Success respond")
}

// GetServices godoc
// @Summary      Get all services
// @Description  Retrieves a list of all service instances.
// @Tags         services
// @Accept       json
// @Produce      json
// @Success      200  {array}  services.Service  "Successfully retrieved the list of services"
// @Failure      500  {object}  map[string]any   "Internal server error"
// @Router       / [get]
func (h *SubscriptionHandler) GetServices(c *gin.Context) {
	srvs, err := h.subscriptionService.GetServices(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get services", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Failed to get services")
		return
	}

	c.JSON(http.StatusOK, srvs)
	logrus.WithFields(logrus.Fields{
		"path":   c.Request.URL.Path,
		"status": http.StatusOK,
		"body":   srvs,
	}).Debug("Success respond")
}

type UpdateRequest struct {
	ServiceName *string    `json:"service_name,omitempty" example:"My Service"`
	Price       *int       `json:"price,omitempty" example:"500"`
	UserID      *uuid.UUID `json:"user_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	StartDate   *time.Time `json:"start_date,omitempty" example:"01-2024"`
	EndDate     *time.Time `json:"end_date,omitempty" example:"12-2025"`
}

// UpdateService godoc
// @Summary      Update an existing service
// @Description  Updates an existing service instance by its name. The fields in the request body are optional and only the provided fields will be updated.
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        ID  		   path  string         true  "UUID of the service to update" example:"123e4567-e89b-12d3-a456-426614174009"
// @Param        request       body  UpdateRequest  true  "Request update service with optional fields"
// @Success      200           {object} services.Service  "Successfully updated the service"
// @Failure      400           {object} map[string]any    "Invalid UUID or request body or malformed data"
// @Failure      404           {object} map[string]any    "Service not found"
// @Failure      500           {object} map[string]any    "Internal server error"
// @Router       /{id} [patch]
func (h *SubscriptionHandler) UpdateService(c *gin.Context) {
	ID, err := uuid.Parse(c.Param("ID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Invalid UUID")
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Invalid request body")
		return
	}

	updatedSrv := &services.Service{ID: ID}
	if req.ServiceName != nil {
		updatedSrv.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		updatedSrv.Price = *req.Price
	}
	if req.UserID != nil {
		updatedSrv.UserID = *req.UserID
	}
	if req.StartDate != nil {
		updatedSrv.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		updatedSrv.EndDate = req.EndDate
	}

	if err := h.subscriptionService.UpdateService(c, updatedSrv); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			logrus.WithFields(logrus.Fields{
				"path":    c.Request.URL.Path,
				"details": err.Error(),
			}).Warn("Service not found")
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update service", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Failed to update service")
		return
	}

	c.JSON(http.StatusOK, updatedSrv)
	logrus.WithFields(logrus.Fields{
		"path":   c.Request.URL.Path,
		"status": http.StatusOK,
		"body":   updatedSrv,
	}).Debug("Success respond")
}

// DeleteService godoc
// @Summary      Delete a service
// @Description  Deletes a service instance by its name.
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        ID  		   path  string         true  "UUID of the service to delete" example:"123e4567-e89b-12d3-a456-426614174009"
// @Success      204           "Successfully deleted the service"
// @Failure      400           {object} map[string]any    "Invalid UUID"
// @Failure      404           {object} map[string]any    "Service not found"
// @Failure      500           {object} map[string]any    "Internal server error"
// @Router       /{id} [delete]
func (h *SubscriptionHandler) DeleteService(c *gin.Context) {
	ID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Invalid UUID")
		return
	}

	if err = h.subscriptionService.DeleteService(c, ID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			logrus.WithFields(logrus.Fields{
				"path":    c.Request.URL.Path,
				"details": err.Error(),
			}).Warn("Service not found")
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update service", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Failed to update service")
		return
	}

	c.JSON(http.StatusOK, nil)
	logrus.WithFields(logrus.Fields{
		"path":   c.Request.URL.Path,
		"status": http.StatusOK,
	}).Debug("Success respond")
}

type CumulateFiltersRequest struct {
	ServiceNames []*string    `json:"service_name,omitempty" example:"[\"My Service\", \"Someone's Service\"]"`
	UserIDs      []*uuid.UUID `json:"user_id,omitempty" example:"[\"123e4567-e89b-12d3-a456-426614174000\"]"`
	StartDate    string       `json:"start_date" example:"01-2024"`
	EndDate      string       `json:"end_date" example:"12-2025"`
}

// CumulateServices godoc
// @Summary      Cumulate service costs
// @Description  Calculates the total cost of services based on provided filters for date range, user ID and service name.
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        filters  body  services.Filters  true  "Filters for cumulation, including date range and optional users ID and service names"
// @Success      200      {number}  float64       	 "Successfully calculated the total cost"
// @Failure      400      {object}  map[string]any   "Invalid request body or filters"
// @Failure      500      {object}  map[string]any   "Internal server error"
// @Router       /cummulate [get]
func (h *SubscriptionHandler) CumulateServices(c *gin.Context) {
	var filtersReq CumulateFiltersRequest
	if err := c.ShouldBindJSON(&filtersReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Invalid request body")
		return
	}

	startDate, err := time.Parse(dateLayout, filtersReq.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Invalid start date")
		return
	}

	endDate, err := time.Parse(dateLayout, filtersReq.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Invalid end date")
		return
	}

	filters := &services.Filters{
		SrvNames:  filtersReq.ServiceNames,
		UserIDs:   filtersReq.UserIDs,
		StartDate: startDate,
		EndDate:   endDate,
	}

	sum, err := h.subscriptionService.CumulateServices(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cumulate price", "details": err.Error()})
		logrus.WithFields(logrus.Fields{
			"path":    c.Request.URL.Path,
			"details": err.Error(),
		}).Warn("Failed to cumulate price")
		return
	}

	c.JSON(http.StatusOK, sum)
	logrus.WithFields(logrus.Fields{
		"path":   c.Request.URL.Path,
		"status": http.StatusOK,
		"body":   sum,
	}).Debug("Success respond")
}
