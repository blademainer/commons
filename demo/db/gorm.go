package main

import (
	"context"
	"fmt"
	dbutil "github.com/blademainer/commons/pkg/db"
	"github.com/blademainer/commons/pkg/generator"
	"github.com/blademainer/commons/pkg/kvcontext"
	"github.com/blademainer/commons/pkg/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // 引入mysql
	"time"
)

type province struct {
	ID   string `json:"id" gorm:"column:id;primary_key"`
	Name string `json:"name" gorm:"column:name"`
}

type city struct {
	ID       string `json:"id" gorm:"column:id;primary_key"`
	Name     string `json:"name" gorm:"column:name"`
	ParentID string `json:"parent_id" gorm:"column:parent_id"`
}

func main() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	db, err := gorm.Open("mysql", "root:admin@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=true&loc=Local&timeout=10s&collation=utf8mb4_unicode_ci")
	if err != nil {
		panic(err)
	}
	logger.SetLevel("debug")
	//db.SetLogger(logger.NewLogger())
	db = db.Debug()

	err = db.Exec("CREATE DATABASE if not exists test DEFAULT CHARSET 'utf8mb4';").Error
	if err != nil {
		fmt.Println(err.Error())
	}
	err = db.CreateTable(&province{}).Error
	if err != nil {
		fmt.Println(err.Error())
	}
	err = db.CreateTable(&city{}).Error
	if err != nil {
		fmt.Println(err.Error())
	}
	defer func() {
		db.DropTableIfExists(&province{})
		db.DropTableIfExists(&city{})
	}()

	g := generator.New("1", 100)
	p := &province{
		ID:   g.GenerateId(),
		Name: "广东省",
	}

	cities := []*city{
		{
			ID:       g.GenerateId(),
			Name:     "深圳市",
			ParentID: p.ID,
		},
		{
			ID:       g.GenerateId(),
			Name:     "东莞市",
			ParentID: p.ID,
		},
		{
			ID:       g.GenerateId(),
			Name:     "广州市",
			ParentID: p.ID,
		},
	}

	err = dbutil.Transaction(ctx, db, func(ctx2 context.Context, tx *gorm.DB) (err error) {
		err = tx.Create(p).Error
		if err != nil {
			return err
		}
		for _, c := range cities {
			err := dbutil.Transaction(ctx2, db, func(ctx context.Context, tx *gorm.DB) (err error) {
				err = tx.Create(c).Error
				return err
			})
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		return nil
	})
	if err != nil {
		panic(err.Error())
	}

	p2 := &province{
		ID:   g.GenerateId(),
		Name: "广东省",
	}

	cities2 := []*city{
		{
			ID:       g.GenerateId(),
			Name:     "深圳市",
			ParentID: p.ID,
		},
		{
			ID:       g.GenerateId(),
			Name:     "东莞市",
			ParentID: p.ID,
		},
		{
			ID:       g.GenerateId(),
			Name:     "广州市",
			ParentID: p.ID,
		},
	}
	fmt.Println(cities2)
	ctx = kvcontext.NewKvContext(ctx)
	err = dbutil.Transaction(ctx, db, func(ctx2 context.Context, tx *gorm.DB) (err error) {
		err = tx.Create(p2).Error
		if err != nil {
			return err
		}
		for _, c := range cities {
			err := dbutil.Transaction(ctx2, db, func(ctx context.Context, tx *gorm.DB) (err error) {
				err = tx.Create(c).Error
				return err
			})
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		return nil
	})
	if err != nil {
		panic(err.Error())
	}
}
