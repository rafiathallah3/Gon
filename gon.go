package gon

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

/*
-- Yang Perlu Dilakukan dalam Web Framework Gon (8/10/2023 18:40)
1. Render - sudah (9/10/2023 19:24)
2. Render Template - sudah (10/10/2023 20:06)
3. Static - sudah (10/10/2023 22:51)
4. MiddleWare - sudah (12/10/2023 21:50)
5. Render JSON - sudah (13/10/2023 17:29)
6. Kue - sudah (14/10/2023 16:40)
7. FuncMap - sudah (14/10/2023 17:45)
8. Flash - sudah (15/10/2023 10:36)
9. Parameter di path - sudah (18/10/2023 17:40)
10. Tambahin Auth
*/

type httpMethod string
type urlPattern string
type TipeDataJson map[string]any
type FuncMap map[string]any
type HandlerFunc func(*Context)

const (
	GET    httpMethod = "GET"
	POST   httpMethod = "POST"
	DELETE httpMethod = "DELETE"
	UPDATE httpMethod = "UPDATE"
)

type SettingCookie struct {
	Path     string
	Domain   string
	Expires  time.Time
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

type routeRules struct {
	methods map[httpMethod]http.Handler
}

type router struct {
	routes             map[urlPattern]routeRules
	middleware_handler http.Handler
	static_path        urlPattern
	fungsi_middleware  []HandlerFunc
	FuncMap            template.FuncMap
	MaxMultipartMemory int64
}

type Context struct {
	Response     http.ResponseWriter
	Request      *http.Request
	SimpananData map[string]any
	router       *router
	apakahNext   bool
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.URL.Path = hilangkanSlahBerlebihan(req.URL.Path)

	var foundRoute routeRules
	var exists bool

	if strings.Contains(req.URL.Path, string(r.static_path)) {
		foundRoute, exists = r.routes[r.static_path]
	} else {
		foundRoute, exists = r.routes[urlPattern(req.URL.Path)]
		/*
			Masalah dalam pembuatan route parameter
			1. /user/:nama/:data/:data1/:data2
			2. /user/:nama/profile || /user/:nama/setting
			3. /user/:nama/setting/:namasetting/test/:namaTest
			4. /:terserah/data
		*/

		// Kemungkinan ini adalah route paramter jadi kita cek semua router, memakai loop itu membuat perfomancenya lambat tapi tidak ada cara lain [17/10/2023 22:42]
		if !exists {
			for nama_route, handler := range r.routes {
				if ParseURLParameter(nama_route, req.URL.Path) == nil {
					continue
				}

				foundRoute = handler
				exists = true

				break
			}
		}

	}

	if strings.ReplaceAll(req.URL.Path, "/", "") == strings.ReplaceAll(string(r.static_path), "/", "") {
		http.NotFound(w, req)
		return
	}

	if !exists {
		http.NotFound(w, req)
		return
	}

	handler, exists := foundRoute.methods[httpMethod(req.Method)]

	if !exists {
		notAllowed(w, req, foundRoute)
		return
	}

	handler.ServeHTTP(w, req)
}

func (router *router) Route(method httpMethod, pattern urlPattern, f HandlerFunc) {
	rules, exists := router.routes[pattern]
	if !exists {
		rules = routeRules{methods: make(map[httpMethod]http.Handler)}
		router.routes[pattern] = rules
	}

	rules.methods[method] = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context := new(Context)
		context.Response = w
		context.Request = r
		context.router = router
		context.SimpananData = map[string]any{}

		if strings.Contains(string(pattern), ":") {
			SplitURL := strings.Split(r.URL.Path, "/")
			for index, karakter := range strings.Split(string(pattern), "/") {
				if strings.Contains(karakter, ":") {
					context.Set(strings.ReplaceAll(karakter, ":", ""), SplitURL[index])
				}
			}
		}

		for _, v := range router.fungsi_middleware {
			v(context)
			if !context.apakahNextFunc() {
				break
			}
		}

		f(context)
	})
}

func (r *router) SetFuncMap(funcMap FuncMap) {
	r.FuncMap = template.FuncMap(funcMap)
}

func (r *router) Use(fungsi HandlerFunc) {
	r.fungsi_middleware = append(r.fungsi_middleware, fungsi)
}

func (r *router) Static(path urlPattern, file_path string) {
	r.static_path = path

	rules := routeRules{methods: make(map[httpMethod]http.Handler)}
	r.routes[path] = rules
	rules.methods[GET] = http.StripPrefix(string(path), http.FileServer(http.Dir(file_path)))
}

func (r *router) Run(port string) {
	fmt.Printf("Server run di: http://localhost%s\n", port)
	http.ListenAndServe(fmt.Sprintf("%s", port), r)
}

func (context *Context) Flash(text string) {
	str := context.dapatinFlash(false)
	str = append(str, text)

	kue := new(http.Cookie)
	kue.Name = "flash"
	kue.Value = encodeBase64([]byte(strings.Join(str, ",")))

	http.SetCookie(context.Response, kue)
}

func (context *Context) dapatinFlash(PakeBatas bool) []string {
	kue, err := context.Request.Cookie("flash")

	if err != nil {
		return nil
	}

	value, err := decodeBase64(kue.Value)

	if err != nil {
		return nil
	}

	str := strings.Split(string(value), ",")

	if PakeBatas {
		kue = new(http.Cookie)
		kue.Name = "flash"
		kue.Value = encodeBase64([]byte(strings.Join(str, ",")))
		kue.MaxAge = -1
		kue.Expires = time.Unix(1, 0)

		http.SetCookie(context.Response, kue)
	}

	return str
}

func (context *Context) Next() {
	context.apakahNext = true
}

func (context *Context) apakahNextFunc() bool {
	if context.apakahNext {
		temp := context.apakahNext
		context.apakahNext = false
		return temp
	}

	return false
}

func (context *Context) Render(isi string) {
	context.Response.Write([]byte(isi))
}

func (context *Context) Redirect(url urlPattern) {
	http.Redirect(context.Response, context.Request, string(url), http.StatusFound)
}

func (context *Context) Render_template(nama_file string, data map[string]interface{}) {
	_, fileErr := os.ReadDir("./pages")
	if fileErr != nil {
		http.Error(context.Response, "No template to render", http.StatusInternalServerError)
		return
	}

	files := []string{}

	filesTemplateHTML, fileErr := os.ReadDir("./pages/templates")
	if fileErr == nil {
		for _, templateHtml := range filesTemplateHTML {
			files = append(files, path.Join("./pages/templates", templateHtml.Name()))
		}
	}

	//Ide yang sangat jelek, supaya bisa include base.htmlnya harus tambahin "nama_file" di akhir files string list, supaya render html nama file tersebut.
	files = append(files, path.Join("./pages", nama_file))

	tmpl, err := template.New(nama_file).Funcs(context.router.FuncMap).ParseFiles(files...)

	if err != nil {
		http.Error(context.Response, err.Error(), http.StatusInternalServerError)
	}

	strFlash := context.dapatinFlash(true)
	data["flashed_messages"] = strFlash

	err = tmpl.Execute(context.Response, data)

	if err != nil {
		http.Error(context.Response, err.Error(), http.StatusInternalServerError)
	}
}

func (context *Context) JSON(dataJson TipeDataJson) {
	hasil, err := json.Marshal(dataJson)

	if err != nil {
		http.Error(context.Response, err.Error(), http.StatusInternalServerError)
		return
	}

	context.Response.Write(hasil)
}

func (context *Context) SetCookie(nama string, value string, setting SettingCookie) {
	kue := new(http.Cookie)

	kue.Name = nama
	kue.Value = value
	kue.Path = setting.Path
	kue.Domain = setting.Domain
	kue.Expires = setting.Expires
	kue.MaxAge = setting.MaxAge
	kue.Secure = setting.Secure
	kue.HttpOnly = setting.HttpOnly

	http.SetCookie(context.Response, kue)
}

func (context *Context) GetCookie(nama string) (*http.Cookie, error) {
	return context.Request.Cookie(nama)
}

func (context *Context) Set(namaVariable string, value any) {
	context.SimpananData[namaVariable] = value
}

func (context *Context) Get(namaVariable string) (any, bool) {
	return context.SimpananData[namaVariable], context.SimpananData[namaVariable] != nil
}

func (context *Context) Query(nama string) string {
	return context.Request.URL.Query().Get(nama)
}

func (context *Context) PostData(nama string) string {
	return context.Request.FormValue(nama)
}

func (context *Context) FormFile(nama string) (*multipart.FileHeader, error) {
	if err := context.Request.ParseMultipartForm(context.router.MaxMultipartMemory); err != nil {
		return nil, err
	}

	file, handler, err := context.Request.FormFile(nama)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	return handler, nil
}

func (context *Context) SaveFile(fileHeader *multipart.FileHeader, dst string) error {
	file, err := fileHeader.Open()

	if err != nil {
		return err
	}

	defer file.Close()

	lokasiFile := path.Join(dst, fileHeader.Filename)
	buatFile, err := os.Create(lokasiFile)
	if err != nil {
		return err
	}

	defer buatFile.Close()

	if _, err := io.Copy(buatFile, file); err != nil {
		return err
	}

	return nil
}

func ParseURLParameter(urlRaw urlPattern, urlAsli string) map[string]string {
	data := map[string]string{}
	splitUrlAsli := strings.Split(urlAsli, "/")
	splitUrlSetting := strings.Split(string(urlRaw), "/")

	if len(splitUrlAsli) != len(splitUrlSetting) {
		return nil
	}

	for index, karakter := range splitUrlSetting {
		if strings.Contains(karakter, ":") {
			data[strings.ReplaceAll(karakter, ":", "")] = splitUrlAsli[index]
			continue
		}

		if karakter != splitUrlAsli[index] {
			return nil
		}
	}

	return data
}

func notAllowed(w http.ResponseWriter, req *http.Request, r routeRules) {
	methods := make([]string, 1)
	for k := range r.methods {
		methods = append(methods, string(k))
	}
	w.Header().Set("Allow", strings.Join(methods, " "))
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func hilangkanSlahBerlebihan(url string) string {
	HasilList := []string{}

	for _, v := range strings.Split(url, "/") {
		if v == "" {
			continue
		}

		HasilList = append(HasilList, v)
	}

	Karakter := "/" + strings.Join(HasilList, "/")
	if Karakter[len(Karakter)-1] == '/' && len(Karakter) > 1 {
		Karakter = Karakter[:len(Karakter)-1]
	}

	return Karakter
}

func encodeBase64(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

func decodeBase64(src string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(src)
}

func New() *router {
	return &router{routes: make(map[urlPattern]routeRules)}
}
