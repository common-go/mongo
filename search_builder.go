package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"strings"
)

type SearchBuilder struct {
	Collection        *mongo.Collection
	ModelType         reflect.Type
	BuildQuery        func(sm interface{}) (bson.M, bson.M)
	BuildSort         func(s string, modelType reflect.Type) bson.M
	ExtractSearchInfo func(m interface{}) (string, int64, int64, int64, error)
	Map               func(ctx context.Context, model interface{}) (interface{}, error)
}

func NewSearchBuilder(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), buildSort func(s string, modelType reflect.Type) bson.M, extract func(m interface{}) (string, int64, int64, int64, error), options...func(context.Context, interface{}) (interface{}, error)) *SearchBuilder {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	collection := db.Collection(collectionName)
	builder := &SearchBuilder{Collection: collection, ModelType: modelType, BuildQuery: buildQuery, BuildSort: buildSort, ExtractSearchInfo: extract, Map: mp}
	return builder
}
func NewSearchBuilderWithMap(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), buildSort func(s string, modelType reflect.Type) bson.M, mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (string, int64, int64, int64, error)) *SearchBuilder {
	var extractSearchInfo func(m interface{}) (string, int64, int64, int64, error)
	if len(options) >= 1 {
		extractSearchInfo = options[0]
	} else {
		extractSearchInfo = ExtractSearchInfo
	}
	collection := db.Collection(collectionName)
	builder := &SearchBuilder{Collection: collection, ModelType: modelType, BuildQuery: buildQuery, BuildSort: buildSort, ExtractSearchInfo: extractSearchInfo, Map: mp}
	return builder
}
func NewSearchBuilderWithQuery(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), options ...func(context.Context, interface{}) (interface{}, error)) *SearchBuilder {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewSearchBuilderWithMap(db, collectionName, modelType, buildQuery, BuildSort, mp, ExtractSearchInfo)
}
func NewDefaultSearchBuilder(db *mongo.Database, collectionName string, modelType reflect.Type, mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (string, int64, int64, int64, error)) *SearchBuilder {
	q := NewQueryBuilder(modelType)
	var extractSearchInfo func(m interface{}) (string, int64, int64, int64, error)
	if len(options) >= 1 {
		extractSearchInfo = options[0]
	} else {
		extractSearchInfo = ExtractSearchInfo
	}
	return NewSearchBuilderWithMap(db, collectionName, modelType, q.BuildQuery, BuildSort, mp, extractSearchInfo)
}
func (b *SearchBuilder) Search(ctx context.Context, m interface{}) (interface{}, int64, error) {
	query, fields := b.BuildQuery(m)

	var sort = bson.M{}
	s, pageIndex, pageSize, firstPageSize, err := b.ExtractSearchInfo(m)
	if err != nil {
		return nil, 0, err
	}
	sort = b.BuildSort(s, b.ModelType)
	return BuildSearchResult(ctx, b.Collection, b.ModelType, query, fields, sort, pageIndex, pageSize, firstPageSize, b.Map)
}
func BuildSearchResult(ctx context.Context, collection *mongo.Collection, modelType reflect.Type, query bson.M, fields bson.M, sort bson.M, pageIndex int64, pageSize int64, initPageSize int64, mp func(context.Context, interface{}) (interface{}, error)) (interface{}, int64, error) {
	optionsFind := options.Find()
	optionsFind.Projection = fields
	if initPageSize > 0 {
		if pageIndex == 1 {
			optionsFind.SetSkip(0)
			optionsFind.SetLimit(initPageSize)
		} else {
			optionsFind.SetSkip(pageSize*(pageIndex-2) + initPageSize)
			optionsFind.SetLimit(pageSize)
		}
	} else {
		optionsFind.SetSkip(pageSize * (pageIndex - 1))
		optionsFind.SetLimit(pageSize)
	}
	if sort != nil {
		optionsFind.SetSort(sort)
	}

	databaseQuery, er0 := collection.Find(ctx, query, optionsFind)
	if er0 != nil {
		return nil, 0, er0
	}

	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	results := reflect.New(modelsType).Interface()
	er1 := databaseQuery.All(ctx, results)
	if er1 != nil {
		return results, 0, er1
	}
	options := options.Count()
	count, er2 := collection.CountDocuments(ctx, query, options)
	if er2 != nil {
		return results, 0, er2
	}
	if mp == nil {
		return results, count, nil
	}
	r2, er3 := MapModels(ctx, results, mp)
	return r2, count, er3
}

func BuildSort(s string, modelType reflect.Type) bson.M {
	var sort = bson.M{}
	if len(s) == 0 {
		return sort
	}
	sorts := strings.Split(s, ",")
	for i := 0; i < len(sorts); i++ {
		sortField := strings.TrimSpace(sorts[i])
		fieldName := sortField
		c := sortField[0:1]
		if c == "-" || c == "+" {
			fieldName = sortField[1:]
		}

		columnName := GetBsonNameForSort(modelType, fieldName)
		sortType := GetSortType(c)
		sort[columnName] = sortType
	}
	return sort
}
