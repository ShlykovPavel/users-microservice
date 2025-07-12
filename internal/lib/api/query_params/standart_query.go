package query_params

import (
	"fmt"
	"github.com/go-playground/validator"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
)

type ListUsersParams struct {
	Search string `validate:"omitempty"`
	Limit  int    `validate:"gte=1,lte=100"`
	Offset int    `validate:"gte=0"`
	Page   int    `validate:"gte=1"`
	Sort   string
}

func ParseStandardQueryParams(query url.Values, log *slog.Logger) (ListUsersParams, error) {
	log = slog.With(
		slog.String("query", query.Encode()),
		slog.String("operation", "ParseListUsersParams"))

	params := ListUsersParams{
		Limit:  10,
		Offset: 0,
		Page:   1,
		Sort:   "id:asc",
	}

	params.Search = query.Get("search")

	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			log.Warn("Invalid limit", "error", err)
			return params, fmt.Errorf("invalid limit parameter: %w", err)
		}
		params.Limit = limit
	}

	if pageStr := query.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Warn("Invalid page", "error", err)
			return params, fmt.Errorf("invalid page format: %s", err)
		}
		params.Page = page
		params.Offset = (page - 1) * params.Limit
	} else if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			log.Warn("Invalid offset", "error", err)
			return ListUsersParams{}, fmt.Errorf("invalid offset format: %w", err)
		}
		params.Offset = offset

	}

	if sortStr := query.Get("sort"); sortStr != "" {
		parts := strings.Split(sortStr, ":")
		if len(parts) != 2 {
			return params, fmt.Errorf("invalid sort format: %s", sortStr)
		}
		field, order := parts[0], parts[1]
		validSortFields := []string{"id", "first_name", "last_name", "email"}
		if !contains(validSortFields, field) {
			return params, fmt.Errorf("invalid sort field: %s", field)
		}
		if order != "asc" && order != "desc" {
			return params, fmt.Errorf("invalid sort order: %s", order)
		}
		params.Sort = sortStr
	}

	if err := validator.New().Struct(&params); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		log.Error("Error validating request body", "err", validationErrors)
		return params, fmt.Errorf("error validating query params: %w", err)
	}

	return params, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
