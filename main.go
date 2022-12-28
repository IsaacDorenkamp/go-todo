package main

import (
	"fmt"
	"log"
	"os"
	// "net/http"
)

// Handle errors that occur
func handle_err(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Test
func test_db() {
	// Test Creation
	incomplete, err := CreateTodo("Test Item", false)
	handle_err(err)
	complete, err := CreateTodo("Test Item 2", true)
	handle_err(err)

	// Test Reading
	ic_model, err := ReadTodo(incomplete.rowid)
	fmt.Printf("Model A: %v | %v | %v\n", ic_model.task, ic_model.complete, ic_model.rowid)
	handle_err(err)
	if ic_model.complete {
		log.Fatal("Model A | Expected incomplete todo, found complete todo")
	}

	c_model, err := ReadTodo(complete.rowid)
	fmt.Printf("Model B: %v | %v | %v\n", c_model.task, c_model.complete, c_model.rowid)
	handle_err(err)
	if !c_model.complete {
		log.Fatal("Model B | Expected complete todo, found incomplete todo")
	}

	// Test Update
	ic_model.complete = true
	err = ic_model.Update()
	handle_err(err)

	ic_model, err = ReadTodo(ic_model.rowid)
	handle_err(err)

	if ic_model.complete != true {
		log.Fatal("Model A | Expected complete todo, found incomplete todo")
	}

	// Test Delete
	c_model_id := c_model.rowid // Will be reset after Delete
	err = c_model.Delete()
	handle_err(err)

	should_be_nil, err := ReadTodo(c_model_id)
	if should_be_nil != nil {
		log.Fatal("Model B | Should be deleted, but non-nil model was returned")
	}

	fmt.Println("All tests passed")
}

func run_test() {
	setup_db(true)
	defer cleanup_db(true)

	test_db()
}

func run_main(port int, devMode bool) {
	setup_db(false)
	defer cleanup_db(false)

	err := SetupRoutes()
	if err != nil {
		log.Fatal(err)
	}

	server := CreateServer(port, devMode)
	log.Printf("Listening on port %v!\n", port)
	server.ListenAndServe()
}

func main() {
	var mode string
	if len(os.Args) > 1 {
		mode = os.Args[1]
	} else {
		mode = "main"
	}

	switch mode {
	case "test":
		run_test()
	case "dev":
		run_main(8080, true)
	case "main":
		run_main(8080, false)
	default:
		log.Fatal(fmt.Sprintf("Invalid mode '%v'", mode))
	}
}