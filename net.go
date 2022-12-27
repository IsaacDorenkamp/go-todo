package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"regexp"
	"encoding/json"
	"strconv"
)

type TodoServer struct {}

type Response struct {
	status int
	content []byte
}

type Route struct {
	path *regexp.Regexp
	handler func(req *http.Request, params []string) *Response
}

func create_json(data map[string]string) []byte {
	output, _ := json.Marshal(data)
	return output
}

func create_error(err string) []byte {
	source := map[string]string {
		"message": err,
	}

	return create_json(source)
}

func create_success(msg string) []byte {
	source := map[string]string {
		"message": msg,
	}

	return create_json(source)
}

func RootTodo(req *http.Request, params []string) *Response {
	method := req.Method
	switch(method) {
	case "GET":
		todos, err := ListTodos()
		if err != nil {
			return &Response {
				status: 500,
				content: create_error("Failed to fetch todos"),
			}
		}

		output, err := json.Marshal(todos)
		if err != nil {
			return &Response {
				status: 500,
				content: create_error("Could not convert todos to JSON format"),
			}
		}

		return &Response {
			status: 200,
			content: output,
		}
	case "POST":
		req.ParseForm()
		// TODO: Refactor for efficiency
		task := req.PostFormValue("task")
		complete_s := req.PostFormValue("complete")
		if task == "" || complete_s == "" {
			return &Response {
				status: 400,
				content: create_error("Request must contain 'task' and 'complete' fields."),
			}
		} else {
			complete, err := strconv.ParseBool(complete_s)
			if err != nil {
				return &Response {
					status: 400,
					content: create_error("'complete' must be a valid boolean."),
				}
			}

			tx, err := GlobalCtx.BeginTx()
			if err != nil {
				log.Println(err)
				return &Response {
					status: 500,
					content: create_error("Could not create todo in database."),
				}
			}

			todo, err := CreateTodoTx(task, complete, tx)
			if err != nil {
				log.Println(err)
				return &Response {
					status: 500,
					content: create_error("Could not create todo in database."),
				}
			}

			content := map[string]any {
				"entity": todo,
				"message": "Todo successfully created.",
			}
			result, err := json.Marshal(content)

			if err != nil {
				tx.Rollback()
				return &Response {
					status: 500,
					content: create_error("Could not create todo in database."),
				}
			} else {
				err = tx.Commit()
				if err != nil {
					return &Response {
						status: 500,
						content: create_error("Could not create todo in database."),
					}
				}
			}

			return &Response {
				status: 201,
				content: result,
			}
		}
	default:
		return &Response {
			status: 405,
			content: create_error(fmt.Sprintf("Method not allowed: %v", method)),
		}
	}
}

func TodoEntity(req *http.Request, params []string) *Response {
	var todo_id int64
	if len(params) == 1 {
		base_todo_id, err := strconv.Atoi(params[0])
		if err != nil {
			return &Response {
				status: 400,
				content: create_error("Did not receive a valid integer ID."),
			}
		} else {
			todo_id = int64(base_todo_id)
		}
	} else {
		return &Response {
			status: 400,
			content: create_error("Did not receive a valid integer ID."),
		}
	}

	todo, err := ReadTodo(todo_id)
	if err != nil {
		return &Response {
			status: 500,
			content: create_error("Could not fetch todo."),
		}
	} else {
		if todo == nil {
			return &Response {
				status: 404,
				content: create_error(fmt.Sprintf("Could not find a todo with ID '%v'.", todo_id)),
			}
		} else {
			// We have our entity, now we can determine what to do with it
			switch (req.Method) {
			case "GET":
				content := map[string]any {
					"entity": todo,
				}
				result, err := json.Marshal(content)
				if err != nil {
					return &Response {
						status: 500,
						content: create_error("Unable to format todo."),
					}
				}

				return &Response {
					status: 200,
					content: result,
				}
			case "PUT":
				req.ParseForm()
				task := req.FormValue("task")
				complete_s := req.FormValue("complete")
				if task == "" || complete_s == "" {
					return &Response {
						status: 400,
						content: create_error("PUT requests must include non-empty 'task' and 'complete' fields."),
					}
				}

				complete, err := strconv.ParseBool(complete_s)
				if err != nil {
					return &Response {
						status: 400,
						content: create_error("'complete' field must be a valid boolean value."),
					}
				}

				tx, err := GlobalCtx.BeginTx()
				if err != nil {
					log.Println(err)
					return &Response {
						status: 500,
						content: create_error("Could not update todo."),
					}
				}

				todo.task = task
				todo.complete = complete
				err = todo.UpdateTx(tx)
				if err != nil {
					tx.Rollback()
					log.Println(err)
					return &Response {
						status: 500,
						content: create_error("Could not update todo."),
					}
				} else {
					content := map[string]any {
						"entity": todo,
						"message": "Update successful.",
					}

					result, err := json.Marshal(content)
					if err == nil {
						err = tx.Commit()
					}

					if err != nil {
						return &Response {
							status: 500,
							content: create_error("Could not update todo."),
						}
					}

					return &Response {
						status: 200,
						content: result,
					}
				}
			case "DELETE":
				tx, err := GlobalCtx.BeginTx()
				if err != nil {
					return &Response {
						status: 500,
						content: create_error("Could not delete todo."),
					}
				}

				del_id := todo.rowid // Store here since it will be reset to -1 upon successful deletion
				err = todo.DeleteTx(tx)

				if err != nil {
					tx.Rollback()
					return &Response {
						status: 500,
						content: create_error("Could not delete todo."),
					}
				} else {
					err = tx.Commit()
					if err != nil {
						return &Response {
							status: 500,
							content: create_error("Could not delete todo."),
						}
					}
					return &Response {
						status: 200,
						content: create_success(fmt.Sprintf("Todo #%v deleted.", del_id)),
					}
				}
			default:
				return &Response {
					status: 405,
					content: create_error(fmt.Sprintf("Method not allowed: %v", req.Method)),
				}
			}
		}
	}
}

func FrontEnd(req *http.Request, _ []string) *Response {

}

func StaticServe(req *http.Request, _ []string) *Response {
	
}

// Route Definitions
var Root []Route
func SetupRoutes() error {
	todo_base, err := regexp.Compile("^/todo$")
	if err != nil {
		return err
	}
	todo_entity, err := regexp.Compile("^/todo/([0-9]+)$")
	if err != nil {
		return err
	}

	base, err := regexp.Compile("^/$")
	if err != nil {
		return err
	}

	static, err := regexp.Compile("^/static/(.+)$")
	if err != nil {
		return err
	}

	Root = []Route {
		Route { path: todo_base, handler: RootTodo },
		Route { path: todo_entity, handler: TodoEntity },
		// Route { path: base, handler: FrontEnd },
		Route { path: static, handler: StaticServe }
	}

	return nil
}

func ResolveRequest(routes []Route, w http.ResponseWriter, req *http.Request) *Response {
	path := req.URL.Path

	var handler func(req *http.Request, params []string) *Response
	var match []string
	for _, candidate := range routes {
		match = candidate.path.FindStringSubmatch(path)
		if match != nil {
			handler = candidate.handler
			break
		}
	}

	if match == nil {
		log.Printf("Could not resolve path: %v", path)
		return nil
	} else {
		log.Printf("%v %v", req.Method, path)
		return handler(req, match[1:])
	}
}

func (server *TodoServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	response := ResolveRequest(Root, w, req)
	if response == nil {
		response = &Response{ status: 404, content: []byte("Not Found"), }
	}

	w.WriteHeader(response.status)
	_, err := w.Write(response.content)
	if err != nil {
		log.Printf("FAILED TO SEND RESPONSE: %v\n", err)
	}
}

func CreateServer(port int) *http.Server {
	handler := &TodoServer{}
	server := &http.Server {
		Addr: fmt.Sprintf(":%v", port),
		Handler: handler,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return server
}