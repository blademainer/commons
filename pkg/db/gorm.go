package db

import (
	"context"
	"database/sql"
	"github.com/blademainer/commons/pkg/kvcontext"
	"github.com/blademainer/commons/pkg/logger"
	"github.com/jinzhu/gorm"
)

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

const (
	// transactionKey is the key for db. Transaction values in Contexts. It is
	// unexported; clients use db.NewContext and db.FromContext
	// instead of using this key directly.
	transactionKey key = iota + 1

	// transactionInvalidKey is the key that marks transaction is invalid.
	transactionInvalidKey
)

// TransactionHandler 需要执行的函数
type TransactionHandler func(ctx context.Context, tx *gorm.DB) (err error)

// Transaction 使用事务处理请求。目前只支持嵌套风格的事务，在退出`f`函数之前将transaction移除掉。示例：
//
//// BatchInsertStatus 一条事务插入task和status
//
//func (s *Storage) BatchInsertStatus(ctx context.Context, task *Task, ss []*Status) error {
//	err := db.Transaction(ctx, s.db, func(ctx context.Context, tx *gorm.DB) (err error) {
// 		err = tx.Create(ctx, task)
// 		if err != nil {
//			return err
//		}
//		for _, status := range ss {
//			err = s.StatusStorage.Create(ctx, status)
//			if err != nil {
//				return err
//			}
//		}
//
//		return nil
//	})
//	return err
//}
//
//// Create create task status // StatusStorage.Create
//
//func (h *StatusStorage) Create(ctx context.Context, r *Status) error {
//	logger := tracinglogger.ContextLog(ctx)
//	logger.Infof("begin Create status, status: %v", r)
//	defer logger.Infof("end Create status")
//
//	return db.Transaction(ctx, h.db, func(ctx context.Context, tx *gorm.DB) error {
//		if len(r.ID) == 0 {
//			r.ID = objectid.New()
//		}
//		r.CreatedTime = utime.NowTimeString()
//		r.UpdatedTime = utime.NowTimeString()
//		err := tx.Create(r).Error
//		return err
//	})
//}
func Transaction(ctx context.Context, db *gorm.DB, f TransactionHandler) (err error) {
	opts := &sql.TxOptions{}
	pt, err := kvcontext.Get(ctx, transactionKey)
	transactionInvalidInterface, err := kvcontext.Get(ctx, transactionInvalidKey)

	transactionInvalid := false
	transactionExists := pt != nil
	//transactionInvalid := false
	// 如果之前没有初始化过KvContext，则初始化kvcontext
	if err != nil && kvcontext.IsNotKvContextError(err) {
		ctx = kvcontext.NewKvContext(ctx)
	}

	// 如果事务未初始化过，则初始化一个
	if !transactionExists {
		pt = db.BeginTx(ctx, opts)
		transactionInvalid = false

		put, err := kvcontext.Put(ctx, transactionKey, pt)
		if err != nil {
			logger.Errorf("failed to put tx, error: %v context: %#v", err.Error(), ctx)
		} else if !put {
			logger.Errorf("put tx error")
		}
		put, err = kvcontext.Put(ctx, transactionInvalidKey, transactionInvalid)
		if err != nil {
			logger.Errorf("failed to put transactionInvalid, error: %v context: %#v", err.Error(), ctx)
		} else if !put {
			logger.Errorf("put transactionInvalid error")
		}
	} else {
		transactionInvalid = transactionInvalidInterface.(bool)
		logger.Debugf("use exists transaction: %v transactionInvalid: %v", pt, transactionInvalidInterface)
	}
	tx := pt.(*gorm.DB)

	// 如果已经被提交过或者已经回滚了就不能再执行下去
	if transactionInvalid {
		logger.Infof("transaction is already invalid")
		return
	}

	defer func() {
		if transactionInvalid {
			logger.Infof("transaction is already invalid")
			return
		}

		if err != nil {
			logger.Errorf("failed to execute sql and returns error: %v, so begin rollback...", err.Error())
			rollback := tx.Rollback()
			if rollback.Error != nil {
				logger.Errorf("failed to rollback, error: %v", rollback.Error.Error())
			} else if logger.IsDebugEnabled() {
				logger.Debugf("successful rollback transaction")
			}

			// 将事务置为不可用
			transactionInvalid = true
			put, err := kvcontext.Put(ctx, transactionInvalidKey, transactionInvalid)
			if err != nil {
				logger.Errorf("failed to put transactionInvalid, error: %v context: %#v", err.Error(), ctx)
			} else if !put {
				logger.Errorf("put transactionInvalid error")
			} else {
				logger.Infof("settings transactionInvalid: %v", transactionInvalid)
			}
		}

		// 如果已经初始化过事务，则该处不应该提交或回滚，而是由最外层的代码来提交，也就是示例里面的：BatchInsertStatus 来提交事务
		if transactionExists {
			return
		}

		commit := tx.Commit()
		if commit.Error != nil {
			logger.Errorf("failed to commit transaction, error: %v", commit.Error.Error())
		} else if logger.IsDebugEnabled() {
			logger.Debugf("successful commit transaction")
		}

		// 将事务置为不可用
		transactionInvalid = true
		put, err := kvcontext.Put(ctx, transactionInvalidKey, transactionInvalid)
		if err != nil {
			logger.Errorf("failed to put transactionInvalid, error: %v context: %#v", err.Error(), ctx)
		} else if !put {
			logger.Errorf("put transactionInvalid error")
		}
	}()

	err = f(ctx, tx)

	transactionInvalidInterface, _ = kvcontext.Get(ctx, transactionInvalidKey)
	if transactionInvalidInterface.(bool) {
		logger.Infof("transaction is already invalid")
		transactionInvalid = true
		return
	}
	return
}

// ReadOnlyTransaction 只读事务
func ReadOnlyTransaction(ctx context.Context, db *gorm.DB, f TransactionHandler) (err error) {
	opts := &sql.TxOptions{ReadOnly: true}
	tx := db.BeginTx(ctx, opts)
	defer func() {
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Errorf("failed to execute sql and returns error: %v, so begin rollback...", err.Error())
			rollback := tx.Rollback()
			if rollback.Error != nil {
				logger.Errorf("failed to rollback, error: %v", rollback.Error.Error())
			}
		} else {
			commit := tx.Commit()
			if commit.Error != nil {
				logger.Errorf("failed to commit transaction, error: %v", commit.Error.Error())
			}
		}
	}()

	err = f(ctx, tx)
	return
}
