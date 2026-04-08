package statistics_transport_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/qandoni/golangToDoApp/internal/core/domain"
	core_logger "github.com/qandoni/golangToDoApp/internal/core/logger"
	core_http_request "github.com/qandoni/golangToDoApp/internal/core/transport/http/request"
	core_http_response "github.com/qandoni/golangToDoApp/internal/core/transport/http/response"
)

type GetStatisticsResponse struct {
	TasksCreated               int      `json:"tasks_created" example:"50"`
	TasksCompleted             int      `json:"tasks_completed" example:"10"`
	TasksCompletedRate         *float64 `json:"tasks_completed_rate" example:"20"`
	TasksAverageCompletionTime *string  `json:"tasks_average_completion_time" example:"1m30s"`
}

// PatchTask godoc
// @Summary Получение статистики
// @Description Изменение статистики по задачами с опциональной фильтрацией по user_id и/или временному промежутку
// @Tags statistics
// @Produce json
// @Param user_id query int false "Фильтрация статистики по конкретному пользователю"
// @Param from query string false "Начало промежутка рассмотрения статистики (включительно)"
// @Param to query int false "Конец промежутку рассмотрения статистики (не включительно)"
// @Success 200 {object} GetStatisticsResponse "Успешно измененная задача"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /statistics [get]
func (h *StatisticsHTTPHandler) GetStatistics(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userIDFromToParams, err := getUserIDFromToQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get UserID/From/To query params",
		)
		return
	}

	statistics, err := h.statisticsService.GetStatistics(
		ctx,
		userIDFromToParams.userID,
		userIDFromToParams.from,
		userIDFromToParams.to,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get statistics",
		)
		return
	}

	response := toDTOFromDomain(statistics)

	responseHandler.JSONResponse(response, http.StatusOK)

}

type queryParams struct {
	userID *int
	from   *time.Time
	to     *time.Time
}

func toDTOFromDomain(statistics domain.Statistics) GetStatisticsResponse {
	var avgTime *string
	if statistics.TasksAverageCompletionTime != nil {
		duration := statistics.TasksAverageCompletionTime.String()
		avgTime = &duration
	}
	return GetStatisticsResponse{
		TasksCreated:               statistics.TasksCreated,
		TasksCompleted:             statistics.TasksCompleted,
		TasksCompletedRate:         statistics.TasksCompletedRate,
		TasksAverageCompletionTime: avgTime,
	}
}

func getUserIDFromToQueryParams(r *http.Request) (queryParams, error) {
	const (
		userIDQueryParamKey = "user_id"
		fromQueryParamKey   = "from"
		toQueryParamKey     = "to"
	)
	userID, err := core_http_request.GetIntQueryParam(r, userIDQueryParamKey)
	if err != nil {
		return queryParams{}, fmt.Errorf("get 'user_id' query param: %w", err)
	}
	from, err := core_http_request.GetDateQueryParam(r, "from")
	if err != nil {
		return queryParams{}, fmt.Errorf("get 'from' query param: %w", err)
	}
	to, err := core_http_request.GetDateQueryParam(r, "to")
	if err != nil {
		return queryParams{}, fmt.Errorf("get 'to' query param: %w", err)
	}
	return queryParams{
		userID: userID,
		from:   from,
		to:     to,
	}, nil
}
