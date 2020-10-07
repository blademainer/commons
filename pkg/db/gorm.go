package db

import (
	"context"
	"database/sql"
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
)

type transactionContext struct {
	tx      *gorm.DB
	invalid bool
}

// TransactionHandler 需要执行的函数
type TransactionHandler func(ctx context.Context, tx *gorm.DB) (err error)

// Transaction 使用事务处理请求。目前只支持嵌套风格的事务，在退出`f`函数之前将transaction移除掉。示例：
//
//// BatchInsertStatus 一条事务插入task和status
//
//func (s *Storage) BatchInsertStatus(ctx context.Context, ss []*Status) error {
//	err := db.Transaction(ctx, s.db, func(ctx context.Context, tx *gorm.DB) (err error) {
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
	select {
	case <-ctx.Done():
		return context.DeadlineExceeded
	default:
	}
	opts := &sql.TxOptions{}

	tc, transactionExists := ctx.Value(transactionKey).(*transactionContext)

	// 如果事务未初始化过，则初始化一个
	if !transactionExists {
		tc = &transactionContext{}

		tc.tx = db.BeginTx(ctx, opts)
		tc.invalid = false

		ctx = context.WithValue(ctx, transactionKey, tc)
	} else {
		logger.Debugf("use exists transactionContext: %#v", tc)
	}

	// 如果已经被提交过或者已经回滚了就不能再执行下去
	if tc.invalid {
		logger.Infof("transactionContext is already invalid")
		return
	}

	defer func() {
		if tc.invalid {
			logger.Infof("transactionContext is already invalid")
			return
		}

		if err != nil {
			logger.Errorf("failed to execute sql and returns error: %v, so begin rollback...", err.Error())
			rollback := tc.tx.Rollback()
			if rollback.Error != nil {
				logger.Errorf("failed to rollback, error: %v", rollback.Error.Error())
			} else if logger.IsDebugEnabled() {
				logger.Debugf("successful rollback transactionContext")
			}

			// 将事务置为不可用
			tc.invalid = true
			ctx = context.WithValue(ctx, transactionKey, tc)
			logger.Infof("settings transactionInvalid: %#v", tc)
		}

		// 如果已经初始化过事务，则该处不应该提交或回滚，而是由最外层的代码来提交，也就是示例里面的：BatchInsertStatus 来提交事务
		if transactionExists {
			return
		}

		commit := tc.tx.Commit()
		if commit.Error != nil {
			logger.Errorf("failed to commit transactionContext, error: %v", commit.Error.Error())
		} else if logger.IsDebugEnabled() {
			logger.Debugf("successful commit transactionContext")
		}

		// 将事务置为不可用
		tc.invalid = true
		ctx = context.WithValue(ctx, transactionKey, tc)
		logger.Infof("settings transactionInvalid: %#v", tc)
	}()

	err = f(ctx, tc.tx)

	tc.invalid = isTransactionInvalid(ctx)
	if tc.invalid {
		logger.Infof("transactionContext is already invalid")
		return
	}
	return
}

func isTransactionInvalid(ctx context.Context) bool {
	tc, transactionExists := ctx.Value(transactionKey).(*transactionContext)
	if !transactionExists {
		return false
	}
	return tc.invalid
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
				logger.Errorf("failed to commit transactionContext, error: %v", commit.Error.Error())
			}
		}
	}()

	err = f(ctx, tx)
	return
}
