package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo"
)

type M map[string]interface{}

type User struct {
	Name  string `json:"name" form:"name" query:"name"`
	Email string `json:"email" form:"email" query:"email"`
}

type User2 struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=80"`
}

type CustomValidator struct {
	validator *validator.Validate
}

var ActionIndex = func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("from action index"))
}

var ActionHome = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("from action index"))
	},
)

var ActionAbout = echo.WrapHandler(
	http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("from action index"))
		},
	),
)

func main() {
	r := echo.New()
	r.GET("/", handler)
	r.GET("/html", handlerHTML)
	r.GET("/redirect", handlerRedirect)
	r.GET("/json", handlerJSON)
	r.GET("/page1", handlerParsing)
	r.GET("/page2/:name", handlerParsing2)
	r.GET("/page3/:name/*", handlerParsing3)
	r.POST("/form", handlerForm)
	r.GET("/index", echo.WrapHandler(http.HandlerFunc(ActionIndex)))
	r.GET("/home", echo.WrapHandler(ActionHome))
	r.GET("/about", ActionAbout)
	r.Any("/user", handlerUser)
	r.POST("/validate", handlerUser2)
	r.Static("/static", "assets")
	r.Validator = &CustomValidator{validator: validator.New()}
	r.HTTPErrorHandler = customError
	//r.HTTPErrorHandler = errorPage
	r.Start(":9000")
}

func handler(ctx echo.Context) error {
	data := "Hello from /index"
	return ctx.String(http.StatusOK, data)
}

func handlerHTML(ctx echo.Context) error {
	data := "Hello from /html"
	return ctx.HTML(http.StatusOK, data)
}

func handlerRedirect(ctx echo.Context) error {
	return ctx.Redirect(http.StatusTemporaryRedirect, "/")
}

func handlerJSON(ctx echo.Context) error {
	data := M{"Message": "Hello", "Counter": 2}
	return ctx.JSON(http.StatusOK, data)
}

func handlerParsing(ctx echo.Context) error {
	name := ctx.QueryParam("name")
	data := fmt.Sprintf("Hello %s", name)

	return ctx.String(http.StatusOK, data)
}

func handlerParsing2(ctx echo.Context) error {
	name := ctx.Param("name")
	data := fmt.Sprintf("Hello %s", name)

	return ctx.String(http.StatusOK, data)
}

func handlerParsing3(ctx echo.Context) error {
	name := ctx.Param("name")
	message := ctx.Param("*")
	data := fmt.Sprintf("Hello %s, I have a message for you: %s", name, message)

	return ctx.String(http.StatusOK, data)
}

func handlerForm(ctx echo.Context) error {
	name := ctx.FormValue("name")
	message := ctx.FormValue("message")

	data := fmt.Sprintf(
		"Hello %s, I have a message for you: %s",
		name,
		strings.Replace(message, "/", "", 1),
	)

	return ctx.String(http.StatusOK, data)
}

func handlerUser(ctx echo.Context) (err error) {
	u := new(User)
	if err = ctx.Bind(u); err != nil {
		return
	}

	return ctx.JSON(http.StatusOK, u)
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func handlerUser2(ctx echo.Context) error {
	u := new(User2)
	if err := ctx.Bind(u); err != nil {
		return err
	}
	if err := ctx.Validate(u); err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, true)
}

func customError(err error, ctx echo.Context) {
	report, ok := err.(*echo.HTTPError)
	if !ok {
		report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if castedObject, ok := err.(validator.ValidationErrors); ok {
		for _, err := range castedObject {
			switch err.Tag() {
			case "required":
				report.Message = fmt.Sprintf("%s is required",
					err.Field())
			case "email":
				report.Message = fmt.Sprintf("%s is not valid",
					err.Field())
			case "gte":
				report.Message = fmt.Sprintf("%s must be greater than %s",
					err.Field(), err.Param())
			case "lte":
				report.Message = fmt.Sprintf("%s must be lower than %s",
					err.Field(), err.Param())
			}

			break
		}
	}

	ctx.Logger().Error(report)
	ctx.JSON(report.Code, report)
}

func errorPage(err error, ctx echo.Context) {
	report, ok := err.(*echo.HTTPError)
	if !ok {
		report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	errPage := fmt.Sprintf("%d.html", report.Code)
	if err := ctx.File(errPage); err != nil {
		ctx.HTML(report.Code, "Errorrrrrrrrr")
	}
}
