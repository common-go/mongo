package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoSearchWriterWithVersionAndSort(db *mongo.Database, collectionName string, modelType reflect.Type, idObjectId bool, version string, buildQuery func(sm interface{}) (bson.M, bson.M), getSort func(interface{}) (string, error), buildSort func(string, reflect.Type) bson.M, options ...Mapper) (*Searcher, *Writer) {
	var mapper Mapper
	if len(options) > 0 && options[0] != nil {
		mapper = options[0]
	}
	if mapper != nil {
		writer := NewWriterWithVersion(db, collectionName, modelType, idObjectId, version, mapper)
		builder := NewSearchBuilderWithSort(db, collectionName, buildQuery, getSort, buildSort, mapper.DbToModel)
		searcher := NewSearcher(builder.Search)
		return searcher, writer
	} else {
		writer := NewWriterWithVersion(db, collectionName, modelType, idObjectId, version)
		builder := NewSearchBuilderWithSort(db, collectionName, buildQuery, getSort, buildSort)
		searcher := NewSearcher(builder.Search)
		return searcher, writer
	}
}

func NewSearchWriterWithVersionAndSort(db *mongo.Database, collectionName string, modelType reflect.Type, version string, buildQuery func(sm interface{}) (bson.M, bson.M), getSort func(interface{}) (string, error), buildSort func(string, reflect.Type) bson.M, options ...Mapper) (*Searcher, *Writer) {
	return NewMongoSearchWriterWithVersionAndSort(db, collectionName, modelType, false, version, buildQuery, getSort, buildSort, options...)
}
func NewSearchWriterWithVersion(db *mongo.Database, collectionName string, modelType reflect.Type, version string, buildQuery func(sm interface{}) (bson.M, bson.M), getSort func(interface{}) (string, error), options ...Mapper) (*Searcher, *Writer) {
	return NewMongoSearchWriterWithVersionAndSort(db, collectionName, modelType, false, version, buildQuery, getSort, BuildSort, options...)
}
func NewSearchWriterWithSort(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), getSort func(interface{}) (string, error), buildSort func(string, reflect.Type) bson.M, options ...Mapper) (*Searcher, *Writer) {
	return NewMongoSearchWriterWithVersionAndSort(db, collectionName, modelType, false, "", buildQuery, getSort, buildSort, options...)
}
func NewSearchWriter(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), getSort func(interface{}) (string, error), options ...Mapper) (*Searcher, *Writer) {
	return NewMongoSearchWriterWithVersionAndSort(db, collectionName, modelType, false, "", buildQuery, getSort, BuildSort, options...)
}
