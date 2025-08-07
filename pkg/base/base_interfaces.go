package base

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// ListOptions provides common filtering, sorting, and pagination options
type ListOptions struct {
	Page      int
	PerPage   int
	SortBy    string
	SortOrder string
	Filters   []FilterOption
	Search    string
}

// FilterOption represents a filter condition for querying
type FilterOption struct {
	Field    string
	Operator string
	Value    interface{}
}

// Operator constants for filtering
const (
	OperatorEqual        = "eq"
	OperatorNotEqual     = "ne"
	OperatorGreaterThan  = "gt"
	OperatorLessThan     = "lt"
	OperatorGreaterEqual = "gte"
	OperatorLessEqual    = "lte"
	OperatorIn           = "in"
	OperatorNotIn        = "not_in"
	OperatorLike         = "like"
	OperatorStartsWith   = "starts_with"
	OperatorEndsWith     = "ends_with"
)

// SortOrder constants
const (
	SortAscending  = "asc"
	SortDescending = "desc"
)

// Validate validates the ListOptions
func (opts *ListOptions) Validate() error {
	var errors []string

	// Validate pagination
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PerPage < 1 {
		opts.PerPage = 10
	}
	if opts.PerPage > 100 {
		opts.PerPage = 100
	}

	// Validate sort order
	if opts.SortOrder != "" &&
		opts.SortOrder != SortAscending &&
		opts.SortOrder != SortDescending {
		errors = append(errors, "invalid sort order")
	}

	// Validate filters
	for _, filter := range opts.Filters {
		if filter.Field == "" {
			errors = append(errors, "filter field cannot be empty")
		}
		if filter.Operator == "" {
			errors = append(errors, "filter operator cannot be empty")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("list options validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// BaseRepository defines common CRUD operations for repositories
type BaseRepository[T any, R any] interface {
	Create(ctx context.Context, value *T) (*R, error)
	FindByID(ctx context.Context, id string) (*R, error)
	Update(ctx context.Context, value *T) (*R, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, opts ListOptions) ([]R, error)
	Count(ctx context.Context, filters []FilterOption) (int, error)
	Exists(ctx context.Context, id string) (bool, error)
	FindByField(ctx context.Context, field string, value interface{}) ([]R, error)
	Search(ctx context.Context, opts ListOptions) ([]R, int, error)

	// Additional methods for bulk operations
	BulkCreate(ctx context.Context, values []*T) ([]R, error)
	BulkUpdate(ctx context.Context, values []*T) ([]R, error)
	BulkDelete(ctx context.Context, ids []string) error
}

// EnhancedRepository extends BaseRepository with additional robust methods
type EnhancedRepository[T any, R any] interface {
	BaseRepository[T, R]

	// Advanced querying methods
	FindByMultipleFields(ctx context.Context, filters map[string]interface{}) ([]R, error)

	// Soft delete and restore operations
	SoftDelete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error

	// Bulk operations with more granular control
	BulkUpsert(ctx context.Context, values []*T) ([]R, error)

	// Transactional operations
	WithTransaction(ctx context.Context, fn func(repo EnhancedRepository[T, R]) error) error
}

// RepositoryOption allows for flexible configuration of repositories
type RepositoryOption[T any] func(interface{}) error

// WithValidation adds validation to repository operations
func WithValidation[T any](validator func(T) error) RepositoryOption[T] {
	return func(repo interface{}) error {
		// Implementation depends on specific repository type
		return nil
	}
}

// WithLogging adds logging to repository operations
func WithLogging[T any](logger interface{}) RepositoryOption[T] {
	return func(repo interface{}) error {
		// Implementation depends on specific repository type
		return nil
	}
}

// BaseService defines common service layer operations
type BaseService[T any, R any] interface {
	Create(ctx context.Context, value *T) (*R, error)
	GetByID(ctx context.Context, id string) (*R, error)
	Update(ctx context.Context, value *T) (*R, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, opts ListOptions) ([]R, error)
	Count(ctx context.Context, filters []FilterOption) (int, error)
	Search(ctx context.Context, query string, opts ListOptions) ([]R, int, error)

	// Additional methods for bulk operations
	BulkCreate(ctx context.Context, values []*T) ([]R, error)
	BulkUpdate(ctx context.Context, values []*T) ([]R, error)
	BulkDelete(ctx context.Context, ids []string) error
}

// BuildFilterOptions helps construct filter options dynamically
func BuildFilterOptions(filters map[string]interface{}) []FilterOption {
	var filterOpts []FilterOption
	for field, value := range filters {
		// Default to equality if no specific operator is needed
		filterOpts = append(filterOpts, FilterOption{
			Field:    field,
			Operator: OperatorEqual,
			Value:    value,
		})
	}
	return filterOpts
}

// IsZero checks if a value is considered zero/empty
func IsZero(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	case reflect.Slice, reflect.Map:
		return rv.Len() == 0
	case reflect.String:
		return rv.String() == ""
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Struct:
		// Special handling for time.Time
		if t, ok := v.(interface{ IsZero() bool }); ok {
			return t.IsZero()
		}
		return false
	default:
		return false
	}
}
