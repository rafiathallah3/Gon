# Gon

Gon is a simple web framework written in Golang, it has alot of feature including: Middleware, Flashing, Sessioning and much more!

# Installation
Install the package by executing this command
```
go get https://github.com/rafiathallah3/Gon
```

# Examples
### Route Example
```go
package main

import "github.com/rafiathallah3/Gon"

func main() {
    server := gon.New()
    
    server.Route("GET", "/", func(context *gon.Context) {
		context.Render("Homepage!")
	})

    server.Run(":8000")
}
```

### Parameter in URL path
Where it must add ":" before the name of the variable
```go
package main

import "github.com/rafiathallah3/Gon"

func main() {
    server := gon.New()
    
    server.Route(gon.GET, "/user/:nama", func(context *gon.Context) {
		Value, exists := context.Get("nama")

        context.Render(Value)
	})

    server.Run(":8000")
}
```

### Function in Template
```go
package main

import "github.com/rafiathallah3/Gon"

func main() {
    server := gon.New()
    
    server.SetFuncMap(gon.FuncMap{
		"getLength": func(s string) int {
			return len(s)
		},
	})
    
    server.Route(gon.GET, "/func", func(context *gon.Context) {
		context.Render_template("funcMap.html", map[string]interface{}{})
	})

    server.Run(":8000")
}
```

Inside funcMap.html
```html
{{define "main"}}
    {{ getLength "Variable" }}
{{end}}
```

### Adding a variable on Context
```go
package main

import "github.com/rafiathallah3/Gon"

func main() {
    server := gon.New()
    
    server.Route(gon.GET, "/set", func(context *gon.Context) {
		context.Set("Variable", "Value")
        context.Render("Setting a value...")
	})

    server.Route(gon.GET, "/get", func(context *gon.Context) {
		Variable, exists := context.Get("Variable")
        
        if !exists {
            context.Render("Variable not exists!")
        } else {
            context.Render("Variable is " + Variable)
        }
	})

    server.Run(":8000")
}
```

### Setting and Getting Cookie
```go
package main

import "github.com/rafiathallah3/Gon"

func main() {
    server := gon.New()
    
    r.Route(gon.GET, "/cookie", func(context *gon.Context) {
        cookie, err := context.GetCookie("cookie")
		context.SetCookie("cookie", "Nyum", gon.SettingCookie{})

		context.Render("cookie Update!")
	})

    server.Run(":8000")
}
```

### Middleware

```go
package main

import "github.com/rafiathallah3/Gon"

func main() {
    server := gon.New()
    
    server.Use(func(ctx *gon.Context) {
		ctx.Set("DataMiddleWare", "First Middleware")
		fmt.Println("FIRST MIDDLE WARE!!")

		ctx.Next()

		fmt.Println("FIRST MIDDLE WARE AFTER NEXT")
	})

    // MULTIPLE MIDDLEWARE!!!!
    server.Use(func(ctx *gon.Context) {
		ctx.Set("DataMiddleWare2", "Second Middleware")
		fmt.Println("SECOND MIDDLE WARE!!")

		ctx.Next()

		fmt.Println("SECOND MIDDLE WARE AFTER NEXT")
	})

    server.Route(gon.GET, "/user/:nama", func(context *gon.Context) {
		Value, exists := context.Get("nama")

        context.Render(Value)
	})

    server.Run(":8000")
}
```

### Rendering a template
To render a template, create a folder called "pages" in your project
```go
package main

import "github.com/rafiathallah3/Gon"

func main() {
    server := gon.New()
    
    server.Route("GET", "/", func(context *gon.Context) {
		context.Render_template("index.html")
	})

    server.Run(":8000")
}
```

To also render a base template, create a folder called "templates", this will include the define function to the render template
Base Template Example
```html
{{define "base.html"}}
<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}{{end}}</title>
</head>

<body>
    {{block "main" .}}{{end}}
</body>

</html>
{{end}}
```

Importing the base template
```html
{{template "base.html"}}

{{define "title"}}
Homepage
{{end}}

{{define "main"}}
<h2>This is a homepage!</h2>
{{end}}
```

### Session
```go
package main

import "github.com/rafiathallah3/Gon"

func main() {
    server := gon.New()
    server.SessionKey = "A@#SADCA@#!ZXCDD" //Session Key must be 16 or 32 length!
    server.SessionPermanent = false

    r.Route("GET", "/login", func(ctx *gon.Context) {
		akun := ctx.GetSession("akun")
		password := ctx.GetSession("password")

		ctx.Render_template("session.html", map[string]interface{}{
			"akun":     akun,
			"password": password,
		})
	})

	r.Route("POST", "/login", func(ctx *gon.Context) {
		ctx.SetSession("akun", ctx.FormData("username"))
		ctx.SetSession("password", ctx.FormData("password"))

		ctx.Redirect("/login")
	})

    server.Run(":8000")
}
```

### Flashing
```go
package main

import "github.com/rafiathallah3/Gon"

func main() {
    server := gon.New()
    
    server.Route("GET", "/flash", func(context *gon.Context) {
		context.Render_template("flash.html", map[string]interface{}{})
	})

	server.Route("POST", "/flash", func(context *gon.Context) {
		context.Flash(str)
		context.Redirect("/flash")
	})

    server.Run(":8000")
}
```

Inside flash.html
```html
{{define "utama"}}
    <h1>Flash!</h1>
    Flash: {{.flashed_messages}}
    {{range $i, $text := .flashed_messages}}
        <span>{{$text}}</span>
    {{end}}
{{end}}
```

### Static 
Static function needs 2 parameters that is, URL path and Folder path 
```go
package main

import "github.com/rafiathallah3/Gon"

func main() {
    server := gon.New()
    
    server.Static("/static", "./static")

    server.Run(":8000")
}
```

### Set Icon
```go
package main

import "github.com/rafiathallah3/Gon"

func main() {
    server := gon.New()
    
    server.SetIcon("./favicon.ico")

    server.Run(":8000")
}
```