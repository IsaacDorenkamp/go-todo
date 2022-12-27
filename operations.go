package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"encoding/json"
	"os"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var ErrNoUpdate = errors.New("No entry was found to update.")
var ErrNoDelete = errors.New("No entry was found to delete.")

type DBContext struct {
	db *sql.DB
}

func (ctx *DBContext) MakeContext(duration time.Duration) context.Context {
	newctx, _ := context.WithTimeout(context.Background(), duration)
	return newctx
}

func (ctx *DBContext) BeginTx() (*sql.Tx, error) {
	return ctx.db.BeginTx(ctx.MakeContext(3 * time.Second), &sql.TxOptions{Isolation: sql.LevelSerializable})
}

type Todo struct {
	task string
	complete bool
	rowid int64
}

func (todo *Todo) MarshalJSON() ([]byte, error) {
	output := map[string]any {
		"task": todo.task,
		"complete": todo.complete,
		"rowid": todo.rowid,
	}

	return json.Marshal(output)
}

// DB Setup
var GlobalCtx *DBContext

func setup_db(test bool) {
	var file string
	if test {
		file = "test.db"
	} else {
		file = "data.db"
	}

	c, err := sql.Open("sqlite3", fmt.Sprintf("file:%v?cache=shared&mode=rwc", file))
	if err != nil {
		log.Fatal(err)
	}

	GlobalCtx = &DBContext { c }

	// check if db is already configured correctly
	result, err := GlobalCtx.db.Query("select name from sqlite_master where type='table' AND name='todo'")
	if !result.Next() {
		// table exists
		_, err = GlobalCtx.db.Exec("create table todo (task text, complete int)")
	}
}

func cleanup_db(test bool) {
	GlobalCtx.db.Close()

	if test {
		os.Remove("./test.db")
	}
}

type DatabaseHandle interface {
	Prepare(string) (*sql.Stmt, error)
	Query(string, ...any) (*sql.Rows, error)
}

// Operations with todos
func CreateTodoTx(task string, complete bool, tx DatabaseHandle) (*Todo, error) {
	create, err := tx.Prepare("insert into todo (task, complete) values (?, ?)")
	if err != nil {
		return nil, err
	}
	var complete_value int
	if complete {
		complete_value = 1
	} else {
		complete_value = 0
	}

	result, err := create.Exec(task, complete_value)
	if err != nil {
		return nil, err
	}
	ins_id, ins_id_err :=  result.LastInsertId()
	if ins_id_err != nil {
		ins_id = -1
	}

	model := &Todo{ task: task, complete: complete, rowid: ins_id }

	return model, err
}

func CreateTodo(task string, complete bool) (*Todo, error) {
	return CreateTodoTx(task, complete, GlobalCtx.db)
}

func ReadTodoCtx(id int64, db DatabaseHandle) (*Todo, error) {
	fetch, err := db.Prepare("select task, complete, rowid from todo where rowid=?")
	if err != nil {
		return nil, err
	}
	result := fetch.QueryRow(id)

	model := &Todo{}
	err = result.Scan(&model.task, &model.complete, &model.rowid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		return model, nil
	}
}

func ReadTodo(id int64) (*Todo, error) {
	return ReadTodoCtx(id, GlobalCtx.db)
}

func ListTodosCtx(db DatabaseHandle) ([]Todo, error) {
	results := make([]Todo, 0)

	todos, err := db.Query("select task, complete, rowid from todo")
	if err != nil {
		return nil, err
	} else {
		defer todos.Close()
		for todos.Next() {
			todo := Todo{}
			err = todos.Scan(&todo.task, &todo.complete, &todo.rowid)
			if err != nil {
				return nil, err
			}
			results = append(results, todo)
		}

		return results, nil
	}
}

func ListTodos() ([]Todo, error) {
	return ListTodosCtx(GlobalCtx.db)
}

func (todo *Todo) UpdateTx(tx DatabaseHandle) error {
	update, err := tx.Prepare("update todo set task=?, complete=? where rowid=?")
	if err != nil {
		return err
	} else {
		result, err := update.Exec(todo.task, todo.complete, todo.rowid)
		if err != nil {
			return err
		} else {
			affected, err := result.RowsAffected()
			if err != nil {
				return err
			}

			if affected != 1 {
				return ErrNoUpdate;
			} else {
				return nil
			}
		}
	}
}

func (todo *Todo) Update() error {
	return todo.UpdateTx(GlobalCtx.db)
}

func (todo *Todo) DeleteTx(tx DatabaseHandle) error {
	delete, err := tx.Prepare("delete from todo where rowid=?")
	if err != nil {
		return err
	} else {
		result, err := delete.Exec(todo.rowid)
		if err != nil {
			return err
		} else {
			affected, err := result.RowsAffected()
			if err != nil {
				return err
			}

			if affected != 1 {
				return ErrNoDelete;
			} else {
				todo.rowid = -1
				return nil
			}
		}
	}
}

func (todo *Todo) Delete() error {
	return todo.DeleteTx(GlobalCtx.db)
}