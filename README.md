Go To-Do
========

Startup
-------

To build the project, simply run the following commands in the project directory:

```
$ npm install
$ npm run build
$ go build
```

Then, simply run `$ ./go-todo` and you will be able to access the UI by navigating
to `localhost:8080` in your browser!

Development
-----------

This project is capable of taking advantage of create-react-app's hot loading
feature. In order to do this, however, you will have to pass an argument into
the Go process to instruct it to run a little differently:

```
$ go run . dev
```

or, if using the binary,

```
$./go-todo dev
```

Then you can use `$ npm start` to run the development server for create-react-app as usual.
To build the React app for static, "production ready" serving, use `$ npm run build`.

If you are wondering what the difference is between the "normal" mode of the server and the
"dev" mode, it is simply a discrepancy caused by cross-origin requests. Even though the React
server and the Go REST API are running on the same host, the difference in ports causes them
to be treated as different origins, and as such, there are heavy restrictions on the API requests
that cause them to fail. The "dev" mode adds headers to all responses that suppress these errors,
but should not be used in a production mode. Although this app is small and doesn't have complex
architecture, it is still good to try to follow best practices!