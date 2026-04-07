package mongo_driver

import (
	"context"
	"g7/common/model_common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	MongoDBName = "game"
)

type MongoDriver struct {
	client *mongo.Client
}

func NewMongoDriver(uri string) (*MongoDriver, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return &MongoDriver{client: client}, nil
}

func (m *MongoDriver) AutoMigrate(model model_common.DBTableInterface) error {
	// MongoDB 无需预定义结构，这里只确保集合被创建
	db := m.client.Database(MongoDBName)

	// 检查集合是否存在
	collections, err := db.ListCollectionNames(context.Background(), map[string]string{"name": model.TableName()})
	if err != nil {
		return err
	}
	if len(collections) > 0 {
		return nil // 已存在
	}

	// 不存在则创建
	err = db.CreateCollection(context.Background(), model.TableName())
	return err
}

func (m *MongoDriver) Insert(model model_common.DBTableInterface) error {
	// collection = 集合名
	// conf_data = 任意结构体
	coll := m.client.Database(MongoDBName).Collection(model.TableName())
	_, err := coll.InsertOne(nil, model)
	return err
}

func (m *MongoDriver) FindOne(model model_common.DBTableInterface, query any) error {
	coll := m.client.Database(MongoDBName).Collection(model.TableName())
	return coll.FindOne(context.Background(), query).Decode(model)
}

func (m *MongoDriver) FindList(result any, query any) error {
	// 注意：这里需要表名，所以传参必须是 DBTables 类型
	// 你可以传一个空模型进去
	if tbl, ok := result.(model_common.DBTableInterface); ok {
		coll := m.client.Database(MongoDBName).Collection(tbl.TableName())
		cur, err := coll.Find(context.Background(), query)
		if err != nil {
			return err
		}
		return cur.All(context.Background(), result)
	}
	return nil
}
