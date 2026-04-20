package mongo_driver

import (
	"context"
	"errors"
	"fmt"
	"g7/common/dbc/dbc_interface"
	"g7/common/model_common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	MongoDBName = "game"
)

type MongoDriver struct {
	client *mongo.Client
	db     *mongo.Database
	tx     mongo.Session // 事务会话
}

func (m *MongoDriver) BatchInsert(models []model_common.DBTableInterface) error {
	//TODO implement me
	panic("implement me")
}

func NewMongoDriver(uri string) (*MongoDriver, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	db := client.Database(MongoDBName)
	return &MongoDriver{client: client, db: db}, nil
}

func (m *MongoDriver) AutoMigrate(model model_common.DBTableInterface) error {
	// MongoDB 无需预定义结构，这里只确保集合被创建

	// 检查集合是否存在
	collections, err := m.db.ListCollectionNames(context.Background(), map[string]string{"name": model.TableName()})
	if err != nil {
		return err
	}
	if len(collections) > 0 {
		return nil // 已存在
	}

	// 不存在则创建
	err = m.db.CreateCollection(context.Background(), model.TableName())
	return err
}

func (m *MongoDriver) Insert(model model_common.DBTableInterface) error {
	// collection = 集合名
	// conf_data = 任意结构体
	coll := m.client.Database(MongoDBName).Collection(model.TableName())
	_, err := coll.InsertOne(nil, model)
	return err
}

func (m *MongoDriver) Update(model model_common.DBTableInterface, updates any, query any, args ...any) error {
	panic("implement MongoDriver not Exec function")
	return nil
}

func (m *MongoDriver) Exec(sql string) error {
	// collection = 集合名
	// conf_data = 任意结构体
	panic("implement MongoDriver not Exec function")
	return nil
}

func (m *MongoDriver) FindOne(model model_common.DBTableInterface, query any) error {
	coll := m.client.Database(MongoDBName).Collection(model.TableName())
	return coll.FindOne(context.Background(), query).Decode(model)
}

func (m *MongoDriver) FindList(result any, query any, params ...any) error {
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

func (m *MongoDriver) FindListPro(table any, query any, order string, size, page int) error {
	panic("implement MongoDriver not FindListPro function")
}

func (m *MongoDriver) IsTableExists(tableName string) bool {
	// 拿到当前数据库
	// 列出集合名
	collections, err := m.db.ListCollectionNames(context.Background(), bson.M{"name": tableName})
	if err != nil {
		return false
	}
	// 长度 > 0 表示存在
	return len(collections) > 0
}

// TxBegin 开启MongoDB事务
func (m *MongoDriver) TxBegin() dbc_interface.DBInterface {
	session, err := m.client.StartSession()
	if err != nil {
		return &MongoDriver{tx: nil}
	}
	_ = session.StartTransaction()
	return &MongoDriver{tx: session, db: m.db, client: m.client}
}

func (m *MongoDriver) TxBatchMQInsert(models []model_common.DBMqInterface) error {
	if m.tx == nil {
		return errors.New("transaction is nil")
	}
	if len(models) == 0 {
		return nil
	}

	docs := make([]interface{}, len(models))
	for i, v := range models {
		docs[i] = v
	}

	coll := m.db.Collection(models[0].TableName())
	_, err := coll.InsertMany(context.Background(), docs)
	return err
}

func (m *MongoDriver) TxCommit() error {
	if m.tx == nil {
		return fmt.Errorf("no transaction")
	}
	return m.tx.CommitTransaction(context.Background())
}

func (m *MongoDriver) TxRollback() error {
	if m.tx == nil {
		return fmt.Errorf("no transaction")
	}
	return m.tx.AbortTransaction(context.Background())
}
