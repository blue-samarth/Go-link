package utils

import (
	"context"
	"math"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// PaginationParams holds query parameters for pagination
// Can be extended to support sorting, filtering
// e.g. via query struct binding in Gin or Fiber

type PaginationParams struct {
	Page  int `form:"page" json:"page"`
	Limit int `form:"limit" json:"limit"`
}

type PaginationMeta struct {
	CurrentPage int `json:"currentPage"`
	TotalPages  int `json:"totalPages"`
	PageSize    int `json:"pageSize"`
	TotalItems  int64 `json:"totalItems"`
}

type PaginatedResponse[T any] struct {
	Message    string         `json:"message"`
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

func normalizePaginationParams(params *PaginationParams) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 10
	}
}

// PaginateSQL paginates a GORM query and fills the response struct.
func PaginateSQL[T any](ctx context.Context, db *gorm.DB, params PaginationParams) (PaginatedResponse[T], error) {
	normalizePaginationParams(&params)

	var total int64
	if err := db.WithContext(ctx).Count(&total).Error; err != nil {
		return PaginatedResponse[T]{}, err
	}

	offset := (params.Page - 1) * params.Limit
	var records []T
	if err := db.WithContext(ctx).
		Limit(params.Limit).
		Offset(offset).
		Find(&records).
		Error; err != nil {
		return PaginatedResponse[T]{}, err
	}

	return PaginatedResponse[T]{
		Message: "Success",
		Data:    records,
		Pagination: PaginationMeta{
			CurrentPage: params.Page,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.Limit))),
			PageSize:    params.Limit,
			TotalItems:  total,
		},
	}, nil
}

// PaginateMongo paginates a MongoDB query.
func PaginateMongo[T any](ctx context.Context, collection *mongo.Collection, filter interface{}, params PaginationParams, opts ...*options.FindOptions) (PaginatedResponse[T], error) {
	normalizePaginationParams(&params)

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return PaginatedResponse[T]{}, err
	}

	finalOpts := options.MergeFindOptions(opts...)
	if finalOpts == nil {
		finalOpts = options.Find()
	}
	finalOpts.SetSkip(int64((params.Page - 1) * params.Limit))
	finalOpts.SetLimit(int64(params.Limit))

	cursor, err := collection.Find(ctx, filter, finalOpts)
	if err != nil {
		return PaginatedResponse[T]{}, err
	}
	defer cursor.Close(ctx)

	var results []T
	for cursor.Next(ctx) {
		var doc T
		if err := cursor.Decode(&doc); err != nil {
			return PaginatedResponse[T]{}, err
		}
		results = append(results, doc)
	}
	if err := cursor.Err(); err != nil {
		return PaginatedResponse[T]{}, err
	}

	return PaginatedResponse[T]{
		Message: "Success",
		Data:    results,
		Pagination: PaginationMeta{
			CurrentPage: params.Page,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.Limit))),
			PageSize:    params.Limit,
			TotalItems:  total,
		},
	}, nil
}
